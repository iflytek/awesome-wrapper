package main

import (
	"encoding/json"
)

func getParams() ([]byte, []byte) {
	query, _ := json.Marshal(map[string]string{"cmd": "health"})
	header, _ := json.Marshal(map[string]string{"method": "GET"})
	logger.Printf("cmdServer.cmd:%s,method:%s\n", "health", "GET")
	return query, header
}
