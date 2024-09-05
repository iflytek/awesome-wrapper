#!/usr/bin/env python
# coding:utf-8
import os.path
import json
import os
import queue
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


class RequestInfo:
    def __init__(self, sid: str, params: dict, user_tag: str = "", status: int = 0):
        self.handle = str(uuid.uuid4().hex)
        self.sid = sid
        self.user_tag = user_tag
        self.params = params
        self.requests = []
        self.stop_q = queue.Queue()
        self.status = status


class PromptInferenceInfo:
    def __init__(self, wrapper,
                 thread_id: str,
                 prompt: str,
                 requestInfo: RequestInfo,
                 result_q: queue.Queue = None):
        self.wrapper = wrapper
        self.requestInfo = requestInfo
        self.thread_id = thread_id
        self.prompt = prompt
        self.request_id = str(uuid.uuid4().hex)
        self.result_q = result_q


def get_payload(reqData: DataListCls):
    return json.loads(reqData.get('input').data.decode('utf-8'))


def resp_data(status, text):
    data = ResponseData()
    data.key = "result"
    data.setDataType(DataText)
    data.status = status
    data.setData(json.dumps(text).encode("utf-8"))
    return data


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
                task.wrapper.inference(task)
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


# 定义服务推理逻辑
class Wrapper(WrapperBase):
    serviceId = "atp"
    version = "v1"

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        import logging
        self.logger = getFileLogger(level=logging.DEBUG)
        self.request_map: dict[str, RequestInfo] = {}
        self.request_map_lock = threading.Lock()
        self.thread_pool_size = 32
        self.thread_pool = None
        self.logger.info(f"wrapper constructed")

    def wrapperInit(self, config: {}) -> int:
        size_str = os.environ.get("THREAD_POOL_SIZE", "32")
        self.thread_pool_size = int(size_str)
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
        self.request_map_lock.acquire()
        requestInfo = self.request_map[handle]
        if requestInfo is None:
            self.logger.error(f"can't get this handle: {handle}")
            return -1
        requestInfo.status = req.get('input').status
        self.request_map_lock.release()

        self.logger.debug(f'start wrapperWrite handle {handle}, sid: {requestInfo.sid}')

        prompt = req.get('input').data.decode('utf-8')
        self.logger.debug(f'wrapperWrite prompt:{prompt}, sid:{requestInfo.sid}')

        thread_id = self.thread_pool.alloc_min_thread()
        inferenceInfo = PromptInferenceInfo(self, thread_id=thread_id, prompt=prompt, requestInfo=requestInfo)
        self.thread_pool.put_task(thread_id, inferenceInfo)
        self.request_map_lock.acquire()
        self.request_map[handle].requests.append(inferenceInfo.request_id)
        self.request_map_lock.release()
        self.logger.debug(f'success wrapperWrite handle: {handle}, thread_id: {thread_id}, request_id: {inferenceInfo.request_id}, sid: {requestInfo.sid}')
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

    # 模拟引擎推理过程
    def inference(self, inferenceInfo: PromptInferenceInfo):
        requestInfo = inferenceInfo.requestInfo
        sid = requestInfo.sid

        prompt = inferenceInfo.prompt
        res = Response()
        content = {
            "data": prompt,
            "length": len(prompt)
        }
        res.list = [resp_data(requestInfo.status, content)]
        callback(res, requestInfo.user_tag)
        self.logger.info(f'streaming infer end, status: {requestInfo.status}, sid: {sid}')

        return
