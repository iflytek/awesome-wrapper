package config

import (
	"log"
	"testing"

	"git.iflytek.com/MeteringAuth/ma-sdk-golang/rep/ver"
)

func TestInitCfg(t *testing.T) {
	var err error
	err = InitCfg("http://10.1.87.70:6868", "metrics", "reporter", "ma-client", ver.Version, 1)
	// data, err := ioutil.ReadFile("ma-sdk.toml")
	if err != nil {
		log.Fatal(err)
	}
	// err = initCfg(data)
	log.Println(err)

}
