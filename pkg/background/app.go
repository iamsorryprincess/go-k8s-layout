package background

import "slices"

type Closer interface {
	Close()
}

type BaseApp struct {
	closers []Closer
}

func (a *BaseApp) DeferCloser(closer Closer) {
	a.closers = append(a.closers, closer)
}

func (a *BaseApp) Close() {
	if len(a.closers) == 0 {
		return
	}

	for _, closer := range slices.Backward(a.closers) {
		closer.Close()
	}
}
