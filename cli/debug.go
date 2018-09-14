package cli

import (
	"log"
	"net/http"
	"net/http/httputil"
)

type loggingTransport struct {
	transport http.RoundTripper
}

func (t loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req != nil {
		dump, err := httputil.DumpRequestOut(req, false)
		if err != nil {
			log.Printf("Could not dump request: %s", err)
		}
		log.Printf("[REQUEST] %s %s\n%s", req.Method, req.URL, string(dump))
	}
	res, err := t.transport.RoundTrip(req)
	if res != nil {
		dump, err := httputil.DumpResponse(res, false)
		if err != nil {
			log.Printf("Could not dump response: %s", err)
		}
		log.Printf("[RESPONSE] %s %s\n%s", req.Method, req.URL, string(dump))
	}
	return res, err
}
