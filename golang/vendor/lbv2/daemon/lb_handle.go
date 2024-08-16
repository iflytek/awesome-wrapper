/*
 *实现GetServer接口
 */
package daemon

import (
	"git.xfyun.cn/AIaaS/xsf/server"
	"git.xfyun.cn/AIaaS/xsf/utils"
)

type StrategyInst interface {
	set(...SetInPutOpt) (err LbErr)
	get(...GetInPutOpt) (nBestNodes []string, nBestNodesErr LbErr)
	init(toolbox *xsf.ToolBox)
	serve(in *xsf.Req, span *xsf.Span, toolbox *xsf.ToolBox) (res *utils.Res, err error)
}
type LbHandle struct {
	toolbox  *xsf.ToolBox
	worker   StrategyInst
	strategy StrategyClassify
}

func (lh *LbHandle) Init(toolbox *xsf.ToolBox) (err error) {
	lh.toolbox = toolbox
	strategy, err := lh.toolbox.Cfg.GetInt(BO, STRATEGY)
	if err != nil {
		lh.toolbox.Log.Errorf("toolbox parse config param strategy error:%s", err.Error())
		return
	}

	lh.strategy = StrategyClassify(strategy)
	if lh.strategy.String() == "Unknown" {
		return ErrLbStrategyIsNotSupport
	}
	switch lh.strategy {
	case lic:
		{
			lh.worker = newLic(lh.toolbox)
		}
	case licEx:
		{
			lh.worker = newLicEx(lh.toolbox)
		}
	}

	if err != nil {
		lh.toolbox.Log.Errorf("start strategy:%v error:%s", strategy, err.Error())
	}

	return

}

//todo 上报引擎实例数相关信息
func (lh *LbHandle) reportInstTimer(subRouterType string) {
}

func (lh *LbHandle) Stop() (err error) {
	return
}
