package util

import (
	"fmt"
	"io/ioutil"
	"os"
)

func WriteToTempFile(dir, name string, data []byte) (string, error) {
	f, err := ioutil.TempFile(dir, name)
	if err != nil {
		return "", err
	}

	fn := f.Name()

	if err = writeAllAndClose(f, data); err == nil {
		return fn, nil
	}

	if err1 := os.Remove(fn); err1 != nil {
		err = fmt.Errorf("%s, %s", err.Error(), err1.Error())
	}

	return "", err
}

func writeAllAndClose(f *os.File, data []byte) error {
	err := writeAll(f, data)
	err1 := f.Close()

	if err == nil && err1 == nil {
		return nil
	}

	if err != nil {
		return err
	}

	return err1
}

func writeAll(f *os.File, data []byte) error {
	for v := data; len(v) > 0; {
		n, err := f.Write(v)
		if err == nil {
			return nil
		}

		if n == 0 {
			return err
		}

		v = v[n:]
	}

	return nil
}
