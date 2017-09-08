package shimgo

import (
	"fmt"
	"time"
)

// Brought from here: https://blog.abourget.net/en/2016/01/04/my-favorite-golang-retry-function/
func retry(attempts int, sleep time.Duration, callback func() error) (err error) {
	for i := 0; ; i++ {
		err = callback()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
