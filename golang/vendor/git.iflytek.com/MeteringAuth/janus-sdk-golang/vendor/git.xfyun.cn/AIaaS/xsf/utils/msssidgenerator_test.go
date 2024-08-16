package utils

import (
	"testing"
	"fmt"
)

func TestMsssidgenerator(t *testing.T) {
	var msssidgenerator MssSidGenerator
	msssidgenerator.Init("iat", "127.0.0.1", "hf")
	fmt.Println(msssidgenerator.GenerateSid())
}
