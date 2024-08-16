/*
 * @CopyRight: Copyright 2019 IFLYTEK Inc
 * @Author: jianjiang@iflytek.com
 * @LastEditors: Please set LastEditors
 * @Desc:
 * @createTime: 2020-07-23 16:19:49
 * @modifyTime: 2020-07-24 18:48:53
 */

package calc

type Funcs struct {
	Func  string `json:"func"`
	Time  int64  `json:"time"`
	Count int    `json:"cnt"`
}

type Subs struct {
	Sub   string  `json:"sub"`
	Funcs []Funcs `json:"funcs"`
}

type Msg struct {
	Appid string `json:"appid"`
	Ver   string `json:"ver"`
	Uid   string `json:"uid"`
	Subs  []Subs `json:"subs"`
}

// toml config type
type CalcTomlConfig struct {
	Calc CalcConfig
	Log  LogConfig
}

type LogConfig struct {
	Level        string
	LogFormat    string `toml:"format"`
	LogPath      string `toml:"path"`
	ConsolePrint bool   `toml:"console_print"`
}

type CalcConfig struct {
	QueueSize     int `toml:"queue_size"`
	ConsumeNumber int `toml:"consume_number"`
	Hosts         []string
	Topic         string
	Timeout       int64
	Enable        bool
}
