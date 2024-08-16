package xsf

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"git.xfyun.cn/AIaaS/xsf/utils"
	"strings"
)

type ToolBoxServer struct {
}

func (t *ToolBoxServer) Cmdserver(ctx context.Context, in *utils.Request) (*utils.Response, error) {
	queryMap, headersMap := make(map[string]string), make(map[string]string)
	if err := json.Unmarshal([]byte(in.Query), &queryMap); err != nil {
		return &utils.Response{}, fmt.Errorf("query:%v,err:%v", in.Query, err)
	}

	if err := json.Unmarshal([]byte(in.Headers), &headersMap); err != nil {
		return &utils.Response{}, fmt.Errorf("headers:%v,err:%v", in.Headers, err)
	}
	if strings.ToUpper(headersMap["method"]) != "GET" {
		return &utils.Response{}, errors.New("don't support the method")
	}
	buf := bytes.NewBuffer(nil)
	cmdServerRouter(queryMap["cmd"], queryMap, buf)
	return &utils.Response{Body: buf.String()}, nil
}
