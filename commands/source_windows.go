package commands

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/zetamatta/nyagos/nodos"
	"github.com/zetamatta/nyagos/shell"
)

func findBatch(name string) (string, bool) {
	lowerName := strings.ToLower(name)
	if strings.HasSuffix(lowerName, ".cmd") || strings.HasSuffix(lowerName, ".bat") {
		return name, true
	}
	tmp := name + ".cmd"
	if _, err := os.Stat(tmp); err == nil {
		return tmp, true
	}
	tmp = name + ".bat"
	if _, err := os.Stat(tmp); err == nil {
		return tmp, true
	}
	return "", false
}

func cmdSource(ctx context.Context, cmd Param) (int, error) {
	verbose := ioutil.Discard
	args := cmd.Args()[1:]
	rawargs := cmd.RawArgs()[1:]
	debug := false

	for args != nil && len(args) > 0 && args[0][0] == '-' {
		switch args[0] {
		case "-v":
			verbose = cmd.Err()
		case "-d":
			debug = true
		default:
			return 1, fmt.Errorf("source: %s: unknown option", args[0])
		}
		args = args[1:]
		rawargs = rawargs[1:]
	}
	if args == nil || len(args) < 1 {
		return 1, errors.New("source: too few arguments")
	}
	if !filepath.IsAbs(args[0]) {
		args[0] = nodos.LookPath(shell.LookCurdirOrder, args[0], "NYAGOSPATH")
	}
	if tmp, ok := findBatch(args[0]); ok {
		args[0] = tmp
		return shell.RawSource(rawargs, verbose, debug, cmd.In(), cmd.Out(), cmd.Err(), cmd.DumpEnv())
	}
	if sh, ok := cmd.(*shell.Cmd); ok {
		if err := sh.Source(ctx, args[0]); err != nil {
			return 1, fmt.Errorf("%s: %s", args[0], err.Error())
		}
		return 0, nil
	}
	return 1, errors.New("source: Could not find shell instance")
}
