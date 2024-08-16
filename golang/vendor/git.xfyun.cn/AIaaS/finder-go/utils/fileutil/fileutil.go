package fileutil

import "os"
import "io/ioutil"

func ExistPath(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func GetSystemSeparator() string {
	s := "/"

	if os.IsPathSeparator('\\') {
		s = "\\"
	}

	return s
}

func WriteFile(path string, data []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
