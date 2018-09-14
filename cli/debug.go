package cli

import (
	"log"
	"net/http"
	"os"
)

func wrapLoggingClient(client *http.Client) *http.Client {
	return &http.Client{
		Jar:     client.Jar,
		Timeout: client.Timeout,
		Transport: &loggingTransport{
			transport: client.Transport,
			log:       log.New(os.Stdout, "", log.LstdFlags),
		},
	}
}

type loggingTransport struct {
	transport http.RoundTripper
	log       *log.Logger
}

func (t loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.logRequest(req)
	res, err := t.transport.RoundTrip(req)
	t.logResponse(res)
	return res, err
}

func (t loggingTransport) logRequest(req *http.Request) {
	if req != nil {
		t.log.Printf("<- %s %s", req.Method, req.URL)
		for key, values := range req.Header {
			for _, value := range values {
				t.log.Printf("<- %s: %s", key, value)
			}
		}
	}
}

func (t loggingTransport) logResponse(res *http.Response) {
	if res != nil {
		t.log.Printf("-> %s %s", res.Proto, res.Status)
		for key, values := range res.Header {
			for _, value := range values {
				t.log.Printf("-> %s: %s", key, value)
			}
		}
	}
}
