package request

import (
	"net/http"
)

//go:generate mockgen -source=preparer.go -destination=preparer_mock.go -package=request

// Preparer is the interface that allows to prepare a http.Request
//
// Prepare accepts and possibly modifies a Request (e.g., adding Headers). Implementations
// must ensure to not share or hold per-invocation state since Preparers may be shared and re-used.
type Preparer interface {
	Prepare(*http.Request) (*http.Request, error)
}

// PrepareFunc is a method that implements the Preparer interface.
type PrepareFunc func(*http.Request) (*http.Request, error)

// Prepare implements the Preparer interface on PrepareFunc.
func (f PrepareFunc) Prepare(r *http.Request) (*http.Request, error) {
	return f(r)
}

// CombinePreparer combines multiple preparers into a single one
func CombinePreparer(preparers ...Preparer) Preparer {
	return PrepareFunc(func(req *http.Request) (*http.Request, error) {
		var err error
		for _, preparer := range preparers {
			req, err = preparer.Prepare(req)
			if err != nil {
				return req, err
			}
		}

		return req, nil
	})
}

var NoopPreparer = PrepareFunc(func(req *http.Request) (*http.Request, error) { return req, nil })
