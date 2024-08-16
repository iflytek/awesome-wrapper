package common

type File struct {
	Name string
	Data []byte
}

type ConfigChangedCallback interface {
	ConfigChanged([]*File)
	ConfigAdded([]*File)
	ConfigRemoved([]*File)
}