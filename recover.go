package zapex

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

type RecoveryFunc func(err error)

func OnRecover(f RecoveryFunc) {
	if e := recover(); e != nil {
		err := errors.New(printany(e))
		f(err)
	}
}

func WithRecover(f func()) error {
	var rErr error

	func() {
		defer OnRecover(func(err error) {
			rErr = err
		})

		f()
	}()

	return rErr
}

// Taken from runtime/error.go in the standard library (how it prints panics)
func printany(i interface{}) string {
	switch v := i.(type) {
	case nil:
		return "nil"
	case fmt.Stringer:
		return v.String()
	case error:
		return v.Error()
	case int:
		return strconv.Itoa(v)
	case string:
		return v
	default:
		return fmt.Sprintf("%#v", v)
	}
}
