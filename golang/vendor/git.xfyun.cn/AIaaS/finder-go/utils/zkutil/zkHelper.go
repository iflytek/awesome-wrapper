package zkutil

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/cooleric/curator"
	"github.com/cooleric/go-zookeeper/zk"
)

func getRetryPolicy(maxRetryNum int, maxSleepTime time.Duration) curator.RetryPolicy {
	return curator.NewExponentialBackoffRetry(time.Millisecond, maxRetryNum, maxSleepTime)
}

func onConnectionStateChanged(f curator.CuratorFramework, e curator.ConnectionState) {
	switch e {
	case curator.CONNECTED:
		log.Println(e.String())
	case curator.RECONNECTED:
		log.Println(e.String())
	case curator.SUSPENDED:
		log.Println(e.String())
	case curator.LOST:
		log.Println(e.String())
	}
}

func connect(zm *ZkManager, maxRetryNum int, maxSleepTime, connectionTimeout, sessionTimeout time.Duration) error {
	retryPolicy := getRetryPolicy(maxRetryNum, maxSleepTime)
	zm.zkClient = newZkClientWithOptions(strings.Join(zm.MetaData.ZkAddr, ","), retryPolicy, connectionTimeout, sessionTimeout)
	err := zm.zkClient.Start()
	if err != nil {
		return err
	}
	return nil
}

func close(zm *ZkManager) error {
	return zm.zkClient.Close()
}

func addListeners(zm *ZkManager) {
	connListener := curator.NewConnectionStateListener(onConnectionStateChanged)
	listener := curator.NewCuratorListener(func(c curator.CuratorFramework, e curator.CuratorEvent) error {
		// log.Println("listener type:", e.Type().String())
		// log.Println(e.WatchedEvent())
		// log.Println(e.WatchedEvent().Type)
		// EventNodeCreated:         "EventNodeCreated",
		// EventNodeDeleted:         "EventNodeDeleted",
		// EventNodeDataChanged:     "EventNodeDataChanged",
		// EventNodeChildrenChanged: "EventNodeChildrenChanged",
		// EventSession:             "EventSession",
		// EventNotWatching:         "EventNotWatching",

		// log.Println("e", e == nil)
		// log.Println("e.WatchedEvent()", e.WatchedEvent() == nil)
		// log.Println("e.WatchedEvent().Type", e.WatchedEvent().Type)
		if e == nil {
			log.Println("e is nil")
			return errors.New("CuratorListener:e is nil")
		}
		if e.WatchedEvent() == nil {
			log.Println("e.WatchedEvent() is nil")
			return errors.New("CuratorListener:e.WatchedEvent() is nil")
		}

		switch e.WatchedEvent().Type {
		case zk.EventNodeCreated:
			log.Println("watchevent:", e.WatchedEvent())
		case zk.EventNodeDeleted:
			log.Println("watchevent:", e.WatchedEvent())
		case zk.EventNodeDataChanged:
			log.Println("watchevent:", e.WatchedEvent(), e.Data())
			err := zm.GetNodeDataForCallback(e.Path(), onEventNodeDataChanged)
			if err != nil {
				log.Println(err)
				// todo
			}
		case zk.EventNodeChildrenChanged:
			log.Println("watchevent:", e.WatchedEvent(), e.Children(), e.Data())
			err := zm.GetChildrenForCallback(e.Path(), onEventNodeChildrenChanged)
			if err != nil {
				log.Println(err)
				// todo
			}
		case zk.EventSession:
			log.Println("watchevent:", e.WatchedEvent())
			switch e.WatchedEvent().State {
			case zk.StateExpired:
				zm.expired = true
				log.Println("zk expired by self")
				// if c.BlockUntilConnected() == nil {
				// 	zm.Destroy()
				log.Println("begin recvcer session when expired")
				go func() {
					zm.OnZkSessionExpired()
				}()
				log.Println("recocer session end  when expired")
				// }
				break
			case zk.StateHasSession:
				if zm.expired {
					log.Println("begin recover session when StateHasSession")
					zm.OnZkSessionExpired()
					log.Println("recover session end when StateHasSession")
					zm.expired = false
				}

			}
		}

		return nil
	})

	zm.AddConnectionListener(connListener)
	zm.AddListener(listener)
}

func newZkClient(connString string, maxRetryNum int, maxSleepTime time.Duration) curator.CuratorFramework {
	// these are reasonable arguments for the ExponentialBackoffRetry.
	// the first retry will wait 1 second,
	// the second will wait up to 2 seconds,
	// the third will wait up to 4 seconds.
	retryPolicy := curator.NewExponentialBackoffRetry(time.Millisecond, maxRetryNum, maxSleepTime)

	// The simplest way to get a CuratorFramework instance. This will use default values.
	// The only required arguments are the connection string and the retry policy
	return curator.NewClient(connString, retryPolicy)
}

func newZkClientWithOptions(connString string, retryPolicy curator.RetryPolicy, connectionTimeout, sessionTimeout time.Duration) curator.CuratorFramework {
	// using the CuratorFrameworkBuilder gives fine grained control over creation options.
	builder := &curator.CuratorFrameworkBuilder{
		ConnectionTimeout: connectionTimeout,
		SessionTimeout:    sessionTimeout,
		RetryPolicy:       retryPolicy,
	}

	// return builder.ConnectString(connString).Authorization("digest", []byte("user:pass")).Build()
	return builder.ConnectString(connString).Build()
}
