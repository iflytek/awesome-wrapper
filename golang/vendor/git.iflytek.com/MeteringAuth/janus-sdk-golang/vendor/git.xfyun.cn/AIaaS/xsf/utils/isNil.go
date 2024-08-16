package utils

import "unsafe"

func IsNil(obj interface{}) bool {
	type eFace struct {
		rType unsafe.Pointer
		data  unsafe.Pointer
	}
	if obj == nil {
		return true
	}
	return (*eFace)(unsafe.Pointer(&obj)).data == nil
}
