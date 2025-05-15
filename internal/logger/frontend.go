package logger

import (
	"os"
	"sync"

	"github.com/dagger/dagger/dagql/idtui"
)

var isTTY = sync.OnceValue(func() bool {
	if os.Getenv("TTY") == "0" {
		return false
	}
	return true
})

func NewFrontend() idtui.Frontend {
	if isTTY() {
		return idtui.NewPretty(os.Stdout)
	}
	return idtui.NewPlain(os.Stdout)
}
