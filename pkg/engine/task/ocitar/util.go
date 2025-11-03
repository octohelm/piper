package ocitar

import (
	"net/http"
)

func WithRoundTripperFunc(fn func(req *http.Request, next http.RoundTripper) (*http.Response, error)) func(next http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return &roundTripperFn{
			next: next,
			fn:   fn,
		}
	}
}

type roundTripperFn struct {
	next http.RoundTripper
	fn   func(req *http.Request, next http.RoundTripper) (*http.Response, error)
}

func (r *roundTripperFn) RoundTrip(req *http.Request) (*http.Response, error) {
	return r.fn(req, r.next)
}
