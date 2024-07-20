#!/usr/bin/env python
# coding:utf-8
import enum
import os.path
import json
import os
import queue
import subprocess
import threading
import time
import uuid
from aiges.core.types import *

try:
    from aiges_embed import ResponseData, Response, DataListCls, SessionCreateResponse, callback  # c++
except:
    from aiges.dto import Response, ResponseData, DataListCls, SessionCreateResponse, callback

from aiges.sdk import WrapperBase
from aiges.utils.log import getFileLogger
from openai import OpenAI

DataNone = -1
DataBegin = 0  # 首数据
DataContinue = 1  # 中间数据
DataEnd = 2  # 尾数据


class RequestMode(enum.Enum):
    ONCE = 'once'
    ONCE_ASYNC = 'once_async'
    STREAM = 'stream'


class RequestInfo:
    def __init__(self, sid: str, params: dict, user_tag: str = ""):
        self.handle = str(uuid.uuid4().hex)
        self.sid = sid
        self.user_tag = user_tag
        self.params = params
        self.requests = []
        self.stop_q = queue.Queue()


class PromptInferenceInfo:
    def __init__(self, wrapper,
                 thread_id: str,
                 mode: RequestMode,
                 prompt: str,
                 image: str,
                 requestInfo: RequestInfo,
                 # functions: list,
                 result_q: queue.Queue = None):
        self.wrapper = wrapper
        self.requestInfo = requestInfo
        self.thread_id = thread_id
        self.mode = mode
        self.prompt = prompt
        self.image = image
        # self.functions = functions
        self.request_id = str(uuid.uuid4().hex)
        self.result_q = result_q


def launch_openai_server(mode_path: str, server_port: int, gpu_memory_utilization: float):
    try:
        subprocess.Popen([
            'vllm', 'serve', mode_path, '--port', str(server_port), '--gpu_memory_utilization', str(gpu_memory_utilization)
        ])
    except subprocess.CalledProcessError as e:
        print(f"Failed to start vLLM server: {e}")


def get_payload(reqData: DataListCls):
    return json.loads(reqData.get('input').data.decode('utf-8'))


def resp_content(status, text):
    data = ResponseData()
    data.key = "result"
    data.setDataType(DataText)
    data.status = status
    data.setData(json.dumps(text).encode("utf-8"))
    return data


def get_param_temperature(param):
    return float(param.get('temperature', 0.95))


def get_parma_tok_k(param):
    return int(param.get('top_k', 40))


def get_param_max_new_tokens(param):
    return int(param.get('max_tokens', 2048))


def get_param_repetition_penalty(param):
    return float(param.get('repetition_penalty', 1.1))


class ThreadPool:
    def __init__(self, num_threads, wrapper):
        self.num_threads = num_threads
        self.threads = {}
        self.task_queues = {}
        self.lock = threading.Lock()
        self.wrapper = wrapper
        import logging
        self.filelogger = getFileLogger(level=logging.DEBUG)

        for i in range(num_threads):
            process_id = os.getpid()
            thread_id = "process-{}-thread-{}".format(str(process_id), str(i))
            task_queue = queue.Queue()
            self.threads[thread_id] = threading.Thread(target=self.task_loop, args=(thread_id, task_queue))
            self.task_queues[thread_id] = task_queue

        for thread in self.threads.values():
            thread.start()

    def task_loop(self, thread_id, task_queue):
        self.filelogger.debug(f"task_loop {thread_id} enter")
        while True:
            if not task_queue.empty():
                task: PromptInferenceInfo = task_queue.get_nowait()
                if task is None:
                    break
                task.wrapper.openaiInference(task)
            else:
                time.sleep(0.2)
        self.filelogger.info(f"task_loop {thread_id} end")

    def alloc_min_thread(self) -> str:
        with self.lock:
            min_thread_id = min(self.threads, key=lambda thread_id: self.task_queues[thread_id].qsize())
        return min_thread_id

    def put_task(self, thread_id, task):
        with self.lock:
            self.task_queues[thread_id].put(task)

    def wait_completion(self):
        for task_queue in self.task_queues.values():
            task_queue.put(None)

        for thread in self.threads.values():
            thread.join()


def get_free_port():
    import socket
    # 创建一个临时的socket对象
    temp_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    temp_socket.bind(('localhost', 0))  # 绑定到本地地址的一个随机可用端口
    _, port = temp_socket.getsockname()  # 获取分配的端口号
    temp_socket.close()  # 关闭临时socket
    return port


# 定义服务推理逻辑
class Wrapper(WrapperBase):
    serviceId = "atp"
    version = "v1"

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.gpu_memory_utilization = None
        self.client = None
        self.base_model = None
        self.pretrained_name = None
        import logging
        self.logger = getFileLogger(level=logging.DEBUG)
        self.request_map: dict[str, RequestInfo] = {}
        self.request_map_lock = threading.Lock()
        self.thread_pool_size = 32
        self.thread_pool = None
        self.logger.info(f"openai client wrapper constructed")

    def wait_server_ready(self, server_url: str):
        import requests
        while True:
            try:
                response = requests.get(server_url + "/models", timeout=(1, 3))
                # 检查响应状态码
                if response.status_code == 200:
                    self.logger.info(f"{server_url} ready, {response.content}")
                    break
                else:
                    self.logger.info(f"{server_url} not ready")
                response.close()
            except Exception as e:
                self.logger.info(f"{server_url} connect exception: {e}")
            time.sleep(5)

    def wrapperInit(self, config: {}) -> int:
        self.base_model = os.environ.get("FULL_MODEL_PATH")
        size_str = os.environ.get("THREAD_POOL_SIZE", "32")
        mem_str = os.environ.get("GPU_MEMORY_UTILIZATION", "0.3")
        self.thread_pool_size = int(size_str)
        self.gpu_memory_utilization = float(mem_str)

        self.logger.info(f"base_model: {self.base_model}")
        if not os.path.isdir(self.base_model):
            self.logger.error(f"not find the base_model in FULL_MODEL_PATH")
            return -1

        # 获取服务监听端口
        port = get_free_port()
        # 服务器实际地址
        serverUrl = f"http://127.0.0.1:{port}/v1"
        # 启动服务器进程
        launch_openai_server(self.base_model, port, self.gpu_memory_utilization)
        # 监听服务器是否启动完成
        self.wait_server_ready(serverUrl)
        # 创建openai客户端
        self.client = OpenAI(base_url=serverUrl, api_key="EMPTY")
        self.pretrained_name = self.client.models.list().data[0].id
        self.logger.debug(f'=======model: {self.pretrained_name}============')
        self.thread_pool = ThreadPool(num_threads=self.thread_pool_size, wrapper=self)
        self.logger.info(f'wrapper init success, create thread: {self.thread_pool_size}')
        return 0

    def wrapperFini(self) -> int:
        self.thread_pool.wait_completion()
        return 0

    def wrapperError(self, ret: int) -> str:
        if ret == 100:
            return "no result.."
        return ""

    def wrapperWrite(self, handle: str, req: DataListCls) -> int:
        self.logger.debug(f'start wrapperWrite handle {handle}')
        payload = get_payload(req)
        prompt = payload["text"]
        image_base64 = payload["image"]

        self.request_map_lock.acquire()
        requestInfo = self.request_map[handle]
        if requestInfo is None:
            self.logger.error("can't get this handle:" % handle)
            return -1
        self.request_map_lock.release()

        thread_id = self.thread_pool.alloc_min_thread()
        inferenceInfo = PromptInferenceInfo(self, thread_id=thread_id, mode=RequestMode.STREAM, prompt=prompt,
                                            image=image_base64, requestInfo=requestInfo)
        self.thread_pool.put_task(thread_id, inferenceInfo)
        self.request_map_lock.acquire()
        self.request_map[handle].requests.append(inferenceInfo.request_id)
        self.request_map_lock.release()
        self.logger.debug(
            f'success wrapperWrite handle: {handle}, thread_id: {thread_id}, request_id: {inferenceInfo.request_id}, sid: {requestInfo.sid}')
        return 0

    def wrapperCreate(self, params: {}, sid: str, persId: int = 0, usrTag: str = "") -> SessionCreateResponse:
        self.logger.info(f'start wrapperCreate {params}')

        requestInfo = RequestInfo(sid, params, usrTag)
        self.request_map_lock.acquire()
        self.request_map[requestInfo.handle] = requestInfo
        self.request_map_lock.release()

        s = SessionCreateResponse()
        s.handle = requestInfo.handle
        s.error_code = 0
        self.logger.debug(f'success wrapperCreate, handle: {requestInfo.handle}, sid: {sid}')
        return s

    def wrapperDestroy(self, handle: str) -> int:
        self.logger.debug(f'start wrapperDestroy {handle}')
        self.request_map_lock.acquire()
        requestInfo = self.request_map[handle]
        if requestInfo is None:
            self.logger.error("can't get this handle:" % handle)
            self.request_map_lock.release()
            return -1
        self.request_map_lock.release()
        requestInfo.stop_q.put(True, block=False)
        self.request_map_lock.acquire()
        del self.request_map[handle]
        self.request_map_lock.release()

        self.logger.debug(f'success wrapperDestroy, handle: {handle}, sid: {requestInfo.sid}')
        return 0

    def wrapperTestFunc(self, data: [], respData: []):
        pass

    def openaiInference(self, inferenceInfo: PromptInferenceInfo):
        requestInfo = inferenceInfo.requestInfo
        request_id = inferenceInfo.request_id
        user_tag = requestInfo.user_tag
        sid = requestInfo.sid
        is_stopped = False
        if not requestInfo.stop_q.empty():
            is_stopped = requestInfo.stop_q.get_nowait()
        if is_stopped:
            state = Once
            if inferenceInfo.mode == RequestMode.STREAM:
                state = DataEnd
            content = resp_content(state, ' ')
            res = Response()
            res.list = [content]
            if (inferenceInfo.mode == RequestMode.STREAM or
                    inferenceInfo.mode == RequestMode.ONCE_ASYNC):
                callback(res, user_tag)
            elif inferenceInfo.mode == RequestMode.ONCE:
                inferenceInfo.result_q.put(res)
            self.logger.info(f'====>inference abort before infer, request_id: {request_id}, sid: {sid}')
            return

        params = requestInfo.params
        prompt = inferenceInfo.prompt
        image = inferenceInfo.image
        temperature = get_param_temperature(params)
        max_tokens = get_param_max_new_tokens(params)

        openaiMsgs = [
            {
                "role": "user",
                "content": [
                    {"type": "text", "text": prompt},
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": f"data:image/jpeg;base64,{image}"
                        },
                    },
                ],
            }
        ]

        full_content = ""
        try:
            # 如果是普通流式请求
            result_tokens_len = 0
            streamResp = self.client.chat.completions.create(
                model=self.pretrained_name,
                messages=openaiMsgs,
                temperature=temperature,
                max_tokens=max_tokens,
                stream=True,
            )
            state = DataNone
            for chunk in streamResp:
                if not requestInfo.stop_q.empty():
                    is_stopped = requestInfo.stop_q.get_nowait()
                if is_stopped:
                    # 提前结束
                    self.logger.info(f'====>inference abort when infer, request: {request_id}, sid: {sid}')
                    break
                if chunk.choices[0].delta.content is not None:
                    res = Response()
                    if state == DataNone:
                        state = DataBegin
                    elif state == DataBegin:
                        state = DataContinue
                    content = resp_content(state, chunk.choices[0].delta.content)
                    full_content = full_content + str(chunk.choices[0].delta.content)
                    res.list = [content]
                    callback(res, user_tag)
                    result_tokens_len += 1

            state = DataEnd
            content = resp_content(state, ' ')
            # 流式无法获取输入prompt的tokens数量
            res = Response()
            res.list = [content]
            callback(res, user_tag)
        except Exception as e:
            import traceback
            traceback.print_exc()
            self.logger.error(f"An error occurred when infer: {e}, sid: {sid}")

        self.logger.info(f'streaming infer end, sid: {sid}')

        return
