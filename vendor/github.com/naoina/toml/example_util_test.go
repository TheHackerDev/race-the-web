package toml_test

import (
	"strings"
)

func indent(b []byte, spaces int) string {
	space := strings.Repeat(" ", spaces)
	return space + strings.Replace(string(b), "\n", "\n"+space, -1)
}
