package utils

import "genitus/flange"

type customLogImpl struct{
	flange.CustomLogInterface
	log *Logger
}

func NewLogImpl(l *Logger)*customLogImpl{
   cli := new(customLogImpl)
   cli.log = l
   return cli
}

func (cli *customLogImpl)Infof(format string, params ...interface{}){
    cli.log.Infof(format, params...)
}

func (cli *customLogImpl)Debugf(format string, params ...interface{}){
	cli.log.Debugf(format, params...)
}

func (cli *customLogImpl)Errorf(format string, params ...interface{}){
	cli.log.Errorf(format, params...)
}


func (cli *customLogImpl)Info(params ...interface{}){
	cli.log.Infof("", params...)
}

func (cli *customLogImpl)Debug(params ...interface{}){
	cli.log.Debugf("", params...)
}

func (cli *customLogImpl)Error( params ...interface{}) {
	cli.log.Errorf("", params...)
}

