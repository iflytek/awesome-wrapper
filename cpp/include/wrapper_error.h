#ifndef __WRAPPER_ERROR_H__
#define __WRAPPER_ERROR_H__

enum {
    /* Generic Error defines */  
    Success = 0,
    NotSupportError = 21110,
    CurlInitError = 21111,
    ApiUnauthorized = 21112,
    ApiRequestFailed = 21113,
    ApiResponseFailed = 21114,
    ReqDataEmpty = 21115,
    ReqParamsInvalid = 21116,
    JsonParseError = 21117
};

inline const char *ErrorToString(int err) {
    switch (err) {
        case NotSupportError:
            return "its not support";
        case CurlInitError:
            return "curl init error";
        case ApiUnauthorized:
            return "third part api unauthorized";
        case ApiRequestFailed:
            return "third part api request fail";
        case ApiResponseFailed:
            return "third part api response fail";            
        case ReqDataEmpty:
            return "request data is empty";
        case ReqParamsInvalid:
            return "invalid req params(empty)";
        case JsonParseError:
            return "json parse error";
        default:
            // err.int to err.string ; TODO itoa()
            return "unKnown wrapper error code:";
    }
    return nullptr;
}

#endif /*__WRAPPER_ERROR_H__*/