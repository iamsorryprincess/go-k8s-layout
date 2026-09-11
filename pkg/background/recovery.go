package background

import "fmt"

func RecoverPanic(fatal chan<- error) {
	if r := recover(); r != nil {
		select {
		case fatal <- fmt.Errorf("panic: %v", r):
		default:
		}
	}
}
