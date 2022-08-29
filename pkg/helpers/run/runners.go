package run

import "fmt"

func AsyncWithError(fn func() error) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				if perr, ok := p.(error); ok {
					errCh <- perr
				} else {
					errCh <- fmt.Errorf("panic: %v", p)
				}
			}
		}()

		errCh <- fn()
	}()

	return errCh
}
