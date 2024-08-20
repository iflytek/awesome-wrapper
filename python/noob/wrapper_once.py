#!/usr/bin/env python
# coding:utf-8
import enum
import logging
import json
from aiges.core.types import *

try:
    from aiges_embed import ResponseData, Response, DataListCls, SessionCreateResponse  # c++
except:
    from aiges.dto import Response, ResponseData, DataListCls, SessionCreateResponse

from aiges.sdk import WrapperBase
from aiges.utils.log import getFileLogger


def response_data(text):
    content = {
        "data": text,
        "length": len(text)
    }

    data = ResponseData()
    data.key = "result"
    data.setDataType(DataText)
    data.status = DataOnce
    data.setData(json.dumps(content).encode("utf-8"))
    return data


LOG_LEVEL_MAP = {
    "error": logging.ERROR,
    "warning": logging.WARNING,
    "warn": logging.WARN,
    "info": logging.INFO,
    "debug": logging.DEBUG,
}


# 定义服务推理逻辑
class Wrapper(WrapperBase):
    serviceId = "atp"
    version = "v1"

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.logger = None

    def wrapperInit(self, config: {}) -> int:
        print(f'wrapperInit configs: {config}')
        logLevel = config["log_level"]
        self.logger = getFileLogger(level=LOG_LEVEL_MAP[logLevel])
        return 0

    def wrapperFini(self) -> int:
        return 0

    def wrapperError(self, ret: int) -> str:
        if ret == 100:
            return "no result.."
        return ""

    def wrapperOnceExec(self, params: {}, reqData: DataListCls, usrTag: str = "", persId: int = 0) -> Response:
        sid = params["sid"]
        self.logger.debug(f"WrapperOnceExec, params: {params}, sid: {sid}")

        text = reqData.get('input').data.decode('utf-8')
        self.logger.debug(f'wrapperOnceExec, text: {text}, sid: {sid}')

        res = Response()
        res.list = [response_data(text)]
        return res

    def wrapperTestFunc(self, data: [], respData: []):
        pass
