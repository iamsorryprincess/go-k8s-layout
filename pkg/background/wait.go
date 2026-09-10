package background

import (
	"os"
	"os/signal"
	"syscall"
)

func Wait(fatal <-chan error) (os.Signal, error) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case s := <-exit:
		signal.Stop(exit)
		return s, nil
	case err := <-fatal:
		return nil, err
	}
}
