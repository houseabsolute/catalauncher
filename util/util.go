package util

import (
	"fmt"
	"io"
	"os"
)

func Say(output io.Writer, tmpl string, args ...interface{}) {
	output.Write([]byte(maybeFormat(tmpl, args) + "\n"))
}

func maybeFormat(tmpl string, args []interface{}) string {
	if len(args) == 0 {
		return tmpl
	}

	return fmt.Sprintf(tmpl, args...)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
