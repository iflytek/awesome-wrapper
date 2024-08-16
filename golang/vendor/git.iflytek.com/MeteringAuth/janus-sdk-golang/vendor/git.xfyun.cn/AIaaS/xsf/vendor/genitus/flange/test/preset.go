package test

import (
	"genitus/flange"
	"fmt"
	"time"
)

func GetASpan() *flange.Span {
	span := flange.NewSpan(flange.SERVER, false).Next(flange.SERVER).Next(flange.SERVER).WithName("sessionbegin").Start()
	sid := fmt.Sprintf("iam56eb0006@lc%x3319210", time.Now().UnixNano()/1000/1000)
	span.WithTag("sid", sid)
	span.WithTag("appid", "100IME")
	span.WithTag("uid", "v2077407681")
	span.WithRetTag("10030")
	span.WithTag("cancel", "success")
	span.WithTag("{\"format\" : \"json\"}", "{\"appid\":\"4f0a594c\"}")
	span.WithErrorTag("no server response")
	span.WithDescf("hello %s", "world").WithDescf("hello %d", 3)
	span.End()
	return span
}
