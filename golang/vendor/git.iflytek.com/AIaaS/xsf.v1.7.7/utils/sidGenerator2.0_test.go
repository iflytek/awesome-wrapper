package utils

import (
	"fmt"
	"testing"
)

func TestSidGenerator2_NewSid(t *testing.T) {
	sidGenerator := SidGenerator2{}
	sidGenerator.Init("hf", "127.0.0.1", "9999")
	for i := 0; i < 10; i++ {
		fmt.Println(sidGenerator.NewSid("ctm"))
	}
}
