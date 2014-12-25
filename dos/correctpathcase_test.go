package dos

import (
	"os"
	"strings"
	"testing"
)

func testFixPathCase(t *testing.T, path string) {
	newpath, err := CorrectPathCase(path)
	if err != nil {
		t.Errorf("CorrectPathCase: %v", err)
	}
	println(path, "->", newpath)
}

func TestFixPathCase(t *testing.T) {
	path1, err1 := os.Getwd()
	if err1 != nil {
		t.Errorf("os.Getwd(): %v", err1)
	}
	path1 = strings.ToUpper(path1)
	testFixPathCase(t, path1)
	testFixPathCase(t, "c:\\")
}
