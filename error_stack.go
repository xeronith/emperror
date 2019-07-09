package emperror

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// GetStackTrace returns the stack trace from an error (if any).
func GetStackTrace(err error) (errors.StackTrace, bool) {
	st, ok := getStackTracer(err)
	if ok {
		return st.StackTrace(), true
	}

	return nil, false
}

// getStackTracer returns the stack trace from an error (if any).
func getStackTracer(err error) (stackTracer, bool) {
	var st stackTracer

	UnwrapEach(err, func(err error) bool {
		if s, ok := err.(stackTracer); ok {
			st = s

			return false
		}

		return true
	})

	return st, st != nil
}

type withExposedStack struct {
	err error
	st  stackTracer
}

func (w *withExposedStack) Error() string {
	return w.err.Error()
}

func (w *withExposedStack) Cause() error  { return w.err }
func (w *withExposedStack) Unwrap() error { return w.err }

func (w *withExposedStack) StackTrace() errors.StackTrace {
	return w.st.StackTrace()
}

// Format implements the fmt.Formatter interface.
func (w *withExposedStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = fmt.Fprintf(s, "%+v", w.Cause())
			return
		}
		fallthrough

	case 's':
		_, _ = io.WriteString(s, w.Error())

	case 'q':
		_, _ = fmt.Fprintf(s, "%q", w.Error())
	}
}

// ExposeStackTrace exposes the stack trace (if any) in the outer error.
func ExposeStackTrace(err error) error {
	if err == nil {
		return err
	}

	st, ok := getStackTracer(err)
	if !ok {
		return err
	}

	return &withExposedStack{
		err: err,
		st:  st,
	}
}
