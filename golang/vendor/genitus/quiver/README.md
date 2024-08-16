# Quiver Integration Specification

Quiver, the server/component integrated instrumentation, produces the `event` data and then transfer them to Kafka. We will show you how to use it in this file.

## How to use?

1. download or checkout this repo, extract it into your local desk;
2. add the root directory into your local `GOPATH`;
3. code as the `specification` and run;

## Specification

<font color=red size=72>!!! DON'T USE QUIVER-INST ASYNCHRONOUSLY !!!</font>

* Initialization: to use quiver, you need initialize it before sending anything;

```go
// initialize quiver at the entry of your component
defer quiver.Fini()		// you may not need to call this function

quiver.DumpEnable = true
quiver.DeliverEnable = false
quiver.Logger = &quiver.FmtLog{}  // init with logger inst you use
// set spill dir
quiver.SpillDir = "./log/spill"
// set the flume ip:port, and backend-consumer number inside
quiver.Init("127.0.0.1", "4545", 4,
    "4YBRUQ786P0CVF05WA0B", "pidI9T6Jvlovjuj1XHtsULhCMPVQy3PyHjyqwtrN", "http://172.29.71.35:80",
    "172.29.20.78")
```

* Create a Event: to record `event`, you should create a new `EventData` by calling `quiver.NewEventWithNamePort()`;

```go
// create a new event with parameter, such as type, sid, serviceName, port, NOTICE: event data with unvalid `sid`, such as "", will be ignored while flush
eventSSB := quiver.NewEventWithNamePort(quiver.TYPE_IAT, "iat8b4fc7a7@sc15fc77eb86c84103a0", "iat", "9092")
```

* Add Properties: after create a new event, you can add some properties to record;

```go
// add the properties value
eventSSB.WithUid("v1042331810").
  WithSyncId(0).
  WithSub("iat").
  WithName("ssb").
  WithEndpoint("172.26.5.200")

// update the sid
eventSSB.WithSid("iat8b4fc7a7@sc15fc77eb86c84103a0")

// add a tags key-value
eventSSB.Tag(quiver.KV{"login_id", "v1042331810@100IME"})
// also add other tags
eventSSB.Tag(quiver.KV{"is_open", true}).
  Tag(quiver.KV{"socker_id", 8477}).
  Tag(quiver.KV{"nginx_ip", "172.27.131.11:26559"})

// your can add the `ds` tag for `Druid` DataSource, and the default value will be `vagus`
eventSSB.TagDS("lc")

// add desc value
eventSSB.Desc("req=call `msp_user_ent_get`(`1042331810`, `cantonese16k`)").
  Desc("req-resId:50 resType:WFST resId:S1 resType:HMM_16K resId:S2 resType:HMM_16K")

// add desc value with format
eventSSB.Descf("req:%d, res:%s, %f", 50, "WFST", 16.22)

// add output with specified key
eventSSB.Output("list1", "sdfasdfasdasdfsadfsaaaaa").
  Output("list1", "1464d9f6s6f4s64fd6s4fds6fd")

eventSSB.Output("list2", "34623476245724572457427524572457245724").
  Output("list2", "ghsdfhdhgaadfhgadgfadfga5")

// add media value
audio := []byte("audio-123")
text := []byte("text-123")
eventSSB.Media("audio/L16", "speex-wb", "special=xxx,format=xxxx,encoding=speex-wb", audio).
  Media("audio/L16", "utf-8", "special=xxx,format=xxxx,encoding=speex-wb", text)
```

* Flush: after all tags and properties have been added, you can flush this event to data-channel;

```go
// flush this event to data-channel
eventSSB.Flush()
```

## Demo

There is a demo to show how to support a `event` while coding:

```go
package main

import (
  "genitus/quiver"
)

func main() {
  defer quiver.Fini()		// you may not need to call this function

  quiver.Logger = &quiver.FmtLog{}  // init with logger inst you use
  quiver.DumpEnable = true
  quiver.DeliverEnable = false
  quiver.Init("127.0.0.1", "4545", 4,
    "4YBRUQ786P0CVF05WA0B", "pidI9T6Jvlovjuj1XHtsULhCMPVQy3PyHjyqwtrN", "http://172.29.71.35:80",
    "172.29.20.78")

  test2()
}

func test2() {
  // your code ...

  eventSSB := quiver.NewEventWithNamePort(quiver.TYPE_IAT, "iat8b4fc7a7@sc15fc77eb86c84103a0", "iat", "9092").
    WithUid("v1042331810").
    WithSid("iat8b4fc7a7@sc15fc77eb86c84103a0").
    WithSyncId(0).
    WithSub("iat").
    WithName("ssb").
    WithEndpoint("172.26.5.200")

  eventSSB.Tag(quiver.KV{"login_id", "v1042331810@100IME"}).
    Tag(quiver.KV{"is_open", true}).
    Tag(quiver.KV{"socker_id", 8477}).
    Tag(quiver.KV{"params", "sub=iat, auf=audio/L16; rate=16000, ssm=1, cver=5.0.24.1137"}).
    Tag(quiver.KV{"nginx_ip", "172.27.131.11:26559"}).
    TagDS("lc")

  eventSSB.Desc("req=call `msp_user_ent_get`(`1042331810`, `cantonese16k`)").
    Desc("req-resId:50 resType:WFST resId:S1 resType:HMM_16K resId:S2 resType:HMM_16K").
    Descf("req:%d, res:%s, %f", 50, "WFST", 16.22)

  eventSSB.Output("list1", "sdfasdfasdasdfsadfsaaaaa").
    Output("list1", "1464d9f6s6f4s64fd6s4fds6fd")

  eventSSB.Output("list2", "34623476245724572457427524572457245724").
    Output("list2", "ghsdfhdhgaadfhgadgfadfga5")

  data := []byte("data-123")
  text := []byte("text-123")
  eventSSB.Media("audio/L16", "speex-wb", "special=xxx,format=xxxx,encoding=speex-wb", audio).
            Media("audio/L16", "utf-8", "special=xxx,format=xxxx,encoding=speex-wb", text)

  if err := eventSSB.Flush(); err != nil {
    fmt.Println(err.Error())
  }

  // your code ...
}
```

###
event data saved in hbase&oss will follow this format:
`[header] \n [raw_media_bytes]`