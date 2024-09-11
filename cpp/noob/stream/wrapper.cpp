#include "aiges/wrapper.h"
#include <cstring>
#include <iostream>
#include "aiges/wrapper.h"
#include "wrapper_error.h"
#include "spdlog/sinks/rotating_file_sink.h"
#include "spdlog/spdlog.h"
#include "manager.h"

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
  return Success;
}

int WrapperAPI wrapperExecFree(const char *usrTag, pDataList *respData)
{
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
  manager *hdl = new manager(0, "");
  return (void *)hdl;
}

WrapperAPI int wrapperWrite(const void *handle, pDataList reqData)
{
  char* in = static_cast<char*>(reqData->data);
  manager* hdl = (manager *)handle;
  hdl->set_data(in);
  hdl->set_status(reqData->status);
  return Success;
}

WrapperAPI int wrapperRead(const void *handle, pDataList *respData)
{
  manager* hdl = (manager *)handle;
  std::string out = hdl->get_data();
  int status = hdl->get_status();

    // 封装响应
  *respData = new DataList();

  char *dynamicKey = new char[resp_key.size() + 1];
  strcpy(dynamicKey, resp_key.c_str());
  (*respData)->key = dynamicKey;

  char *dynamicData = new char[out.size() + 1];
  strcpy(dynamicData, out.c_str());

  (*respData)->data = (void *)dynamicData;
  (*respData)->len = out.size();
  (*respData)->desc = nullptr;
  (*respData)->next = nullptr;
  (*respData)->type = DataText;
  (*respData)->status = DataStatus(status);
  return Success;
}

WrapperAPI int wrapperDestroy(const void *handle)
{
  delete handle;
  return Success;
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
