package sproxy

import "git.xfyun.cn/AIaaS/finder-go/rebuild/err"

type EventType int

const(
	NodeChanged EventType = iota
	NodeDeleted
	NodeChildrenAdded
	NodeChildrenRemoved
	NodeChildrenChanged
)

type er = err.Error

type Data struct {
	Path string
	Content []byte
}

type Event struct {
	Type EventType
	Datas []Data
}

type SProxy interface {
	GetPath(path string)([]byte,er)
	WatchPath(path string)(<-chan *Event,er)
	GetChildren(pathdir string)([]Data,er)
	WatchChildren(pathdir string)(<-chan *Event,er)
}

