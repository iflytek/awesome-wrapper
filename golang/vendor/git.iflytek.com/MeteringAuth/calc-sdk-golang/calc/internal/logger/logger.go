/**
 * @CopyRight:   Copyright 2020 IFLYTEK Inc
 * @License:     MIT license
 * @Author:      jianjiang@iflytek.com
 * @CreateTime:  2020/7/30 18:03
 */

package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogFormat int8

var Logger *zap.SugaredLogger

func L(cfg ConfigOption) error {
	if Logger != nil {
		return nil
	}
	l, e := NewLogger(cfg)
	if e != nil {
		return e
	}
	Logger = l
	return nil
}

const (
	JSON LogFormat = iota
	CONSOLE
)

type ConfigOption struct {
	Level        string
	LogPath      string
	ConsolePrint bool
	LogFormat    LogFormat
}

func NewLogger(cfg ConfigOption) (*zap.SugaredLogger, error) {
	cg := zap.NewProductionConfig()
	cg.DisableStacktrace = true
	cg.OutputPaths = []string{cfg.LogPath}
	cg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if cfg.ConsolePrint {
		cg.OutputPaths = append(cg.OutputPaths, "stderr")
	}
	cg.Encoding = "console"
	if cfg.LogFormat == JSON {
		cg.Encoding = "json"
	}

	switch cfg.Level {
	case "debug":
		cg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		cg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		cg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		cg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}
	l, err := cg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	return l.Sugar(), err
}

func Flush() error {
	if Logger == nil {
		return nil
	}
	if err := Logger.Sync(); err != nil {
		fmt.Printf("logger sync error : %s\n", err)
		return err
	}
	return nil
}

func Infow(msg string, kv ...interface{}) {
	if Logger == nil {
		return
	}
	Logger.Infow(msg, kv...)
}

func Debugw(msg string, kv ...interface{}) {
	if Logger == nil {
		return
	}
	Logger.Debugw(msg, kv...)
}

func Warnw(msg string, kv ...interface{}) {
	if Logger == nil {
		return
	}
	Logger.Warnw(msg, kv...)
}
func Errorw(msg string, kv ...interface{}) {
	if Logger == nil {
		return
	}
	Logger.Errorw(msg, kv...)
}
func Fatalw(msg string, kv ...interface{}) {
	if Logger == nil {
		return
	}
	Logger.Fatalw(msg, kv...)
}
