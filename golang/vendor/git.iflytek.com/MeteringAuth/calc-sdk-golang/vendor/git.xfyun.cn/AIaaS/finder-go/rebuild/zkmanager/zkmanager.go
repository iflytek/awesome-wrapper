package zkmanager

import (
	"context"
	"git.xfyun.cn/AIaaS/finder-go/rebuild/log"
	"github.com/cooleric/go-zookeeper/zk"
	"time"
)

type watch struct {

}

type zkManager struct {
	conn *zk.Conn
	watches map[string]*watch
	logger log.Logger
}

//func NewZkManager(conn *zk.Conn,)*zkManager{
//
//}

func (m *zkManager)GetPath(path string)([]byte,error){
	data,_,err:=m.conn.Get(path)
	return data,err
}

func (m *zkManager)SetPath(path string,data []byte)error{
	_,err:=m.conn.Set(path,data,0)
	return err
}

func (m *zkManager)watchEvent(c chan <- *Event,event <-chan zk.Event,path string,ctx context.Context)bool{
	var re zk.Event
	select {
	case <-ctx.Done():
		m.logger.Infof("watch event closed| context canceled,path:",path)
		return false
	case eve:=<-event:
		re = eve
	}
	//re := <-event
	//var et EventType
	var res *Event
	switch re.Type {
	case zk.EventNodeDataChanged:
		//et = NodeChanged
		data, err := m.GetPath(path)
		if err != nil {
			m.logger.Errorf("watchPath|get path data error,path:%s,err:%v", path, err)
			return true
		}
		res = &Event{
			Type: NodeChanged,
			Datas: []Data{
				{
					Path:    re.Path,
					Content: data,
				},
			},
		}
	case zk.EventNodeDeleted:
		//et = NodeDeleted
		res = &Event{
			Type: NodeDeleted,
			Datas: []Data{
				{
					Path:    re.Path,
				},
			},
		}
	case zk.EventNodeChildrenChanged:
		data,_,err:=m.conn.Get(re.Path)
		if err != nil{
			m.logger.Errorf("watchChildren| get children error:path:%s,err:%v",re.Path,err)
			return true
		}
		res = &Event{
			Type: NodeChildrenChanged,
			Datas: []Data{
				{
					Path:    re.Path,
					Content:data,
				},
			},
		}
	case zk.EventNodeCreated:
		data,_,err:=m.conn.Get(re.Path)
		if err != nil{
			m.logger.Errorf("watchChildren| get node created error:path:%s,err:%v",re.Path,err)
			return true
		}
		res = &Event{
			Type: NodeCreated,
			Datas: []Data{
				{
					Path:    re.Path,
					Content:data,
				},
			},
		}
	default:
		m.logger.Warnf("handler unknow event:%v,path:",re.Type,re.Path)

	}


	c <-  res
	return  true
}

func (m *zkManager)watchPath(path string,ctx context.Context)([]byte,chan <- *Event,er){
	c:=make(chan *Event,5)

	data,_,event,err:=m.conn.GetW(path)
	if err != nil{
		return nil,nil,newErr(err.Error())
	}

	go func() {
		if !m.watchEvent(c,event,path,ctx){
			return
		}
		for {
			data,_,event,err =m.conn.GetW(path)
			if err != nil{
				m.logger.Errorf("watchPath|getw  error,path:%s,err:%v", path, err)
				time.Sleep(2*time.Second)
				continue
			}
			if !m.watchEvent(c,event,path,ctx){
				return
			}
		}
	}()
	return data,c,nil

}

func (m *zkManager)watchChildren(chiddir string,ctx context.Context)(chan <- *Event,er){
	_,_,event,err:=m.conn.ChildrenW(chiddir)
	if err != nil{
		return nil, newErr(err.Error())
	}
	c:=make(chan *Event,5)

	go func() {
		if !m.watchEvent(c,event,chiddir,ctx){
			return
		}
		for {
			_,_,event,err :=m.conn.ChildrenW(chiddir)
			if err != nil{
				m.logger.Errorf("watchPath|getw  error,path:%s,err:%v", chiddir, err)
				time.Sleep(2*time.Second)
				continue
			}
			if !m.watchEvent(c,event,chiddir,ctx){
				return
			}
		}
	}()
	return c,nil


}