package app

import (
	"fmt"
)

// SafePrintf wraps fmt.Printf and panics if an error is returned.
func SafePrintf(f string, a ...any) {
	var err error

	_, err = fmt.Printf(f, a...)
	if err != nil {
		panic(err)
	}
}
