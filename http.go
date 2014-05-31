package gutils

import (
	"net/http"
	"time"
	"fmt"
)

type TimeoutTransport struct {
	http.Transport
	RoundTripTimeout time.Duration
}
 
type respAndErr struct {
	resp *http.Response
	err error
}
 
type NetTimeoutError struct {
	error
}
 
func (ne NetTimeoutError) Timeout() bool { return true }
 
// If you don't set RoundTrip on TimeoutTransport, this will always timeout at 0
func (t *TimeoutTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	timeout := time.After(t.RoundTripTimeout)
	resp := make(chan respAndErr, 1)
 
	go func() {
		r, e := t.Transport.RoundTrip(req)
		resp <- respAndErr{
			resp: r,
			err: e,
		}
	}()
 
	select {
	case <-timeout:// A round trip timeout has occurred.
		t.Transport.CancelRequest(req)
		return nil, NetTimeoutError{
			error: fmt.Errorf("timed out after %s", t.RoundTripTimeout),
		}
	case r := <-resp: // Success!
		return r.resp, r.err
	}
}

/*
Usage:

client := &http.Client{
	Transport: &TimeoutTransport{
		RoundTripTimeout: 200 * time.Millisecond,
	},
}

req, err := http.NewRequest("GET", "/path/", nil)
resp, err := client.Do(req) // err could be NetTimeoutError
if err != nil {
	if errTO, ok := err.(NetTimeoutError); ok {
		// timeout
	}
	// some other error
}

*/