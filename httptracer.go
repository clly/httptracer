package httptracer

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptrace"
	"time"
)

// Tracer is a struct that's used to return both a map of values as well as
// the IP address that it connected to. It's mostly used for return values
type Tracer struct {
	Timers map[string]time.Duration
	IP     string
}

// Trace can be used to make a request against a web page and returns all time
// that it takes to do the request. It currently is only tested against GET requests
// and ignores all actual output from the server and follows redirects without telling
// anyone
func Trace(req *http.Request) (*Tracer, error) {
	u := req.URL
	var timers = map[string]time.Duration{}
	var tracer = &Tracer{}
	var t0, t1, t2, t3, t4, tlsstart, tlsdone time.Time

	var trace = &httptrace.ClientTrace{
		// DNS Time is t1 - t0
		DNSStart: func(_ httptrace.DNSStartInfo) { t0 = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { t1 = time.Now() },
		ConnectStart: func(_, _ string) {
			if t1.IsZero() {
				// We're connecting to an IP address set this to DNS time
				t1 = time.Now()
			}
		},
		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				log.Fatalf("unable to connect to host %v: %v", addr, err)
			}
			t2 = time.Now()

			tracer.IP = addr
		},
		TLSHandshakeStart: func() { tlsstart = time.Now() },
		TLSHandshakeDone: func(t tls.ConnectionState, err error) {
			if err != nil {
				log.Fatal("Failed to do tls handshake ", err)
			}
			tlsdone = time.Now()
		},
		GotConn:              func(_ httptrace.GotConnInfo) { t3 = time.Now() },
		GotFirstResponseByte: func() { t4 = time.Now() },
	}

	c := http.DefaultClient
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	w := ioutil.Discard
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	io.Copy(w, resp.Body)
	resp.Body.Close()
	t5 := time.Now() // after reading response body to /dev/null

	if t0.IsZero() {
		// we skipped DNS
		t0 = t1
	}

	switch u.Scheme {
	case "http":
		timers["dns"] = t1.Sub(t0)
		timers["connect"] = t3.Sub(t1)
		timers["server"] = t4.Sub(t3)
		timers["transfer"] = t5.Sub(t4)
		timers["total"] = t5.Sub(t0)
	case "https":
		timers["dns"] = t1.Sub(t0)
		timers["connect"] = t3.Sub(t1)
		timers["tls"] = tlsdone.Sub(tlsstart)
		timers["server"] = t4.Sub(t3)
		timers["transfer"] = t5.Sub(t4)
		timers["total"] = t5.Sub(t0)
	}

	tracer.Timers = timers
	return tracer, nil
}
