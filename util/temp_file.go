package util

import (
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

	err1 := os.Remove(fn)

	return "", MultiErrors(err, err1)
}

func writeAllAndClose(f *os.File, data []byte) error {
	err := writeAll(f, data)
	err1 := f.Close()

	return MultiErrors(err, err1)
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
