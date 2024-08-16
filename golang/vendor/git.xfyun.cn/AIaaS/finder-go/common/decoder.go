package finder

import (
	"encoding/binary"

	errors "git.xfyun.cn/AIaaS/finder-go/errors"
)

func DecodeValue(data []byte) (string, []byte, error) {
	var err error
	if len(data) == 0 {
		err = &errors.FinderError{
			Ret:  errors.InvalidParam,
			Func: "DecodeValue",
			Desc: "data is nil",
		}

		return "", nil, err
	}
	if len(data) <= 4 {
		err = &errors.FinderError{
			Ret:  errors.InvalidParam,
			Func: "DecodeValue",
			Desc: "len of data < =4",
		}

		return "", nil, err
	}
	l := binary.BigEndian.Uint32(data[:4])
	if int(l) > (len(data) - 4) {
		err = &errors.FinderError{
			Ret:  errors.InvalidParam,
			Func: "DecodeValue",
			Desc: "invalid data format",
		}

		return "", nil, err
	}
	pushID := string(data[4 : l+4])

	return pushID, data[l+4:], nil
}
