#include "aiges/wrapper.h"
#include <curl/curl.h> // 头文件路径：/usr/include/x86_64-linux-gnu   库文件路径：/usr/lib/x86_64-linux-gnu
#include <cstring>
#include <iostream>
#include "aiges/wrapper.h"
#include "wrapper_error.h"
#include "nlohmann/json.hpp"
#include "spdlog/sinks/rotating_file_sink.h"
#include "spdlog/spdlog.h"

const char *version = "1.0.0";

std::map<std::string, spdlog::level::level_enum> log_level_map = {
    {"trace", spdlog::level::trace},
    {"debug", spdlog::level::debug},
    {"info", spdlog::level::info},
    {"warn", spdlog::level::warn},
    {"err", spdlog::level::err},
    {"critical", spdlog::level::critical},
    {"off", spdlog::level::off},
};

std::string log_level = "debug";
std::string log_path = "/log/app/wrapper.log";
int log_size = 10;  // 日志文件大小, 默认10MB
int log_count = 30; // 日志文件留存个数, 默认30个
std::string resp_key = "result";

std::string getLogDirectory(const std::string &logPath)
{
  // 查找最后一个斜杠的位置
  size_t lastSlashPos = logPath.find_last_of("/\\");

  // 提取目录路径
  if (lastSlashPos != std::string::npos)
  {
    return logPath.substr(0, lastSlashPos);
  }
  else
  {
    return ""; // 如果路径中不包含斜杠，返回空字符串
  }
}

void initlog()
{
  std::string command = "mkdir -p " + getLogDirectory(log_path);
  auto ret = system(command.c_str());
  if (ret != 0)
  {
    spdlog::error("initlog create log directory failed. ret:{}", ret);
  }

  // 设置全局日志级别为 debug
  spdlog::set_level(log_level_map[log_level]);

  // 设置日志格式
  spdlog::set_pattern("[%l] [%Y-%m-%d %H:%M:%S.%e] [%t] %v");

  SPDLOG_TRACE("Some trace message with param {}", {});
  SPDLOG_DEBUG("Some debug message");

  auto file_logger = spdlog::rotating_logger_mt("noob-wrapper", log_path, 1048576 * log_size, log_count, true);
  spdlog::set_default_logger(file_logger);

  // 根据需要决定是否设置日志刷新方式
  spdlog::flush_on(spdlog::level::debug);
}

wrapperMeterCustom meterPtr;
wrapperTraceLog traceLogPtr;
WrapperAPI int wrapperSetCtrl(CtrlType type, void *func)
{
  if (type == CTMeterCustom)
  {
    spdlog::debug("wrapperSetCtrl type:CTMeterCustom");
    meterPtr = (wrapperMeterCustom)func;
  }
  if (type == CTTraceLog)
  {
    spdlog::debug("wrapperSetCtrl type:CTTraceLog");
    traceLogPtr = (wrapperTraceLog)func;
  }
  return 0;
}

WrapperAPI int wrapperInit(pConfig cfg)
{
  // 传入配置
  while (cfg != nullptr)
  {
    if (cfg->key != nullptr && cfg->value != nullptr)
    {
      if (strcmp(cfg->key, "log_level") == 0)
      {
        log_level = std::string(cfg->value);
      }
      if (strcmp(cfg->key, "log_path") == 0)
      {
        log_path = std::string(cfg->value);
      }
      if (strcmp(cfg->key, "log_size") == 0)
      {
        log_size = std::stoi(std::string(cfg->value));
      }
      if (strcmp(cfg->key, "log_count") == 0)
      {
        log_count = std::stoi(std::string(cfg->value));
      }
    }
    cfg = cfg->next;
  }

  initlog();

  spdlog::info("wrapperInit log_level: {}", log_level);
  spdlog::info("wrapperInit log_path: {}", log_path);
  spdlog::info("wrapperInit log_size: {}", log_size);
  spdlog::info("wrapperInit log_count: {}", log_count);

  spdlog::debug("wrapperInit success");
  return Success;
}

int WrapperAPI wrapperExec(const char *usrTag, pParamList params, pDataList reqData, pDataList *respData, unsigned int psrIds[], int psrCnt)
{
  spdlog::debug("wrapperExec start...");
  std::string sid;

  // 获取请求参数：先取出 sid
  pParamList tmpParams = params;
  while (tmpParams != nullptr)
  {
    if (strcmp(tmpParams->key, "sid") == 0)
    {
      sid = std::string(tmpParams->value);
    }
    tmpParams = tmpParams->next;
  }

  // 获取文本
  if (reqData->len == 0)
  {
    spdlog::debug("wrapperExec req data empty, sid:{}", sid);
    return ReqDataEmpty;
  }
  std::string text = std::string((char *)reqData->data, reqData->len);
  spdlog::debug("wrapperExec input text: {} sid: {}", text, sid);

  // json 结果适配：
  nlohmann::json js;
  js["data"] = text;
  js["length"] = reqData->len;

  std::string result = js.dump();
  spdlog::debug("wrapperExec iflyResult:{} sid:{}", result, sid);

  // 封装响应
  *respData = new DataList();

  char *dynamicKey = new char[resp_key.size() + 1];
  strcpy(dynamicKey, resp_key.c_str());
  (*respData)->key = dynamicKey;

  char *dynamicData = new char[result.size() + 1];
  strcpy(dynamicData, result.c_str());
  (*respData)->data = (void *)dynamicData;

  (*respData)->len = result.size();
  (*respData)->desc = nullptr;
  (*respData)->next = nullptr;
  (*respData)->type = DataText;
  (*respData)->status = DataOnce;

  spdlog::debug("wrapperExec finish..., sid:{}", sid);

  return Success;
}

int WrapperAPI wrapperExecFree(const char *usrTag, pDataList *respData)
{ // respData 二级指针
  spdlog::debug("wrapperExecFree start...");
  // std::cout << "wrapperExecFree start..." << std::endl;
  pDataList resultPtr = *respData;
  if (resultPtr != nullptr)
  {
    if (resultPtr->key != nullptr)
    {
      delete[] resultPtr->key;
      resultPtr->key = nullptr;
    }
    if (resultPtr->data != nullptr)
    {
      delete[] static_cast<char *>(resultPtr->data);
      resultPtr->data = nullptr;
    }
    delete resultPtr;
    resultPtr = nullptr;
  }
  spdlog::debug("wrapperExecFree finish...");
  return Success;
}

WrapperAPI int wrapperFini()
{
  return Success;
}

WrapperAPI const char *wrapperVersion() { return version; }

WrapperAPI const char *wrapperError(int err) { return ErrorToString(err); }

WrapperAPI const void *wrapperCreate(const char *usrTag, pParamList params, wrapperCallback cb, unsigned int psrIds[], int psrCnt, int *errNum)
{
  return nullptr;
}

WrapperAPI int wrapperWrite(const void *handle, pDataList reqData)
{
  return NotSupportError;
}

WrapperAPI int wrapperRead(const void *handle, pDataList *respData)
{
  return NotSupportError;
}

WrapperAPI int wrapperDestroy(const void *handle)
{
  return NotSupportError;
}

WrapperAPI const char *wrapperDebugInfo(const void *handle)
{
  return "not support right now";
}

WrapperAPI int wrapperLoadRes(pDataList perData, unsigned int resId)
{
  return NotSupportError;
}

WrapperAPI int wrapperUnloadRes(unsigned int resId)
{
  return NotSupportError;
}

int WrapperAPI wrapperExecAsync(const char *usrTag, pParamList params, pDataList reqData, wrapperCallback callback, int timeout, unsigned int psrIds[], int psrCnt)
{
  return NotSupportError;
}
