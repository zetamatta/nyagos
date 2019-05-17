package shell

import (
	"errors"
	"io"
	"os"
	"strings"
)

// NoClobber is the switch to forbide to overwrite the exist file.
var NoClobber = false

type _Redirecter struct {
	path     string
	isAppend bool
	no       int
	dupFrom  int
	force    bool
}

func newRedirecter(no int) *_Redirecter {
	return &_Redirecter{
		path:     "",
		isAppend: false,
		no:       no,
		dupFrom:  -1}
}

func (r *_Redirecter) FileNo() int {
	return r.no
}

func (r *_Redirecter) DupFrom(fileno int) {
	r.dupFrom = fileno
}

func (r *_Redirecter) SetPath(path string) {
	r.path = path
}

func (r *_Redirecter) SetAppend() {
	r.isAppend = true
}

func (r *_Redirecter) open() (*os.File, error) {
	if r.path == "" {
		return nil, errors.New("_Redirecter.open(): path=\"\"")
	}
	if strings.EqualFold(r.path, "nul") {
		r.path = os.DevNull
	}
	if r.no == 0 {
		return os.Open(r.path)
	} else if r.isAppend {
		return os.OpenFile(r.path, os.O_APPEND|os.O_CREATE, 0666)
	} else if NoClobber && !r.force {
		return os.OpenFile(r.path, os.O_EXCL|os.O_CREATE, 0666)
	} else {
		return os.Create(r.path)
	}
}

type dontCloseHandle struct{}

func (this dontCloseHandle) Close() error {
	return nil
}

func (r *_Redirecter) OpenOn(cmd *Cmd) (closer io.Closer, err error) {
	var fd *os.File

	switch r.dupFrom {
	case 0:
		fd = cmd.Stdio[0]
		closer = &dontCloseHandle{}
	case 1:
		fd = cmd.Stdio[1]
		closer = &dontCloseHandle{}
	case 2:
		fd = cmd.Stdio[2]
		closer = &dontCloseHandle{}
	default:
		fd, err = r.open()
		if err != nil {
			return nil, err
		}
		closer = fd
	}
	switch r.FileNo() {
	case 0:
		cmd.Stdio[0] = fd
	case 1:
		cmd.Stdio[1] = fd
	case 2:
		cmd.Stdio[2] = fd
	default:
		panic("Assertion failed: _Redirecter.OpenAs: r.no not in (0,1,2)")
	}
	return
}
