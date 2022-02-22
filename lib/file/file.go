package file

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Copy(from, to string) error {
	f, e := os.Stat(from)
	if e != nil {
		return e
	}
	if f.IsDir() {
		if list, e := ioutil.ReadDir(from); e == nil {
			for _, item := range list {
				if e = Copy(filepath.Join(from, item.Name()), filepath.Join(to, item.Name())); e != nil {
					return e
				}
			}
		}
	} else {
		p := filepath.Dir(to)
		if _, e = os.Stat(p); e != nil {
			if e = os.MkdirAll(p, 0777); e != nil {
				return e
			}
		}
		file, e := os.Open(from)
		if e != nil {
			return e
		}
		defer file.Close()
		bufReader := bufio.NewReader(file)
		out, e := os.Create(to)
		if e != nil {
			return e
		}
		defer out.Close()
		_, e = io.Copy(out, bufReader)
	}
	return e
}
