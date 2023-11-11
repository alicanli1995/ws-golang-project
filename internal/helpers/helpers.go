package helpers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"golang-observer-project/internal/config"
	"golang-observer-project/internal/models"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptrace"
	"runtime/debug"
	"time"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var app *config.AppConfig
var src = rand.NewSource(time.Now().UnixNano())

// NewHelpers creates new helpers
func NewHelpers(a *config.AppConfig) {
	app = a
}

// RandomString returns a random string of letters of length n
func RandomString(n int) string {
	b := make([]byte, n)

	for i, theCache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			theCache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(theCache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		theCache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// ServerError will display error page for internal server error
func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	_ = log.Output(2, trace)

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	http.ServeFile(w, r, "./ui/static/500.html")
}

func RenderJSON(w http.ResponseWriter, data interface{}) {
	out, _ := json.MarshalIndent(data, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

func ReadJSONBody(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(dst)
	if err != nil {
		return err
	}
	return nil
}

func ComputeTime(url string) *models.ComputeTimes {
	req, _ := http.NewRequest("GET", url, nil)

	var start, connect, dns, tlsHandshake time.Time
	var computeTimes models.ComputeTimes

	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			computeTimes.DNSDone = time.Since(dns)
		},

		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			computeTimes.TLSHandshake = time.Since(tlsHandshake)
		},

		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			computeTimes.ConnectTime = time.Since(connect)
		},

		GotFirstResponseByte: func() {
			computeTimes.FirstByte = time.Since(start)
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		computeTimes.ResponseStatus = 0
	} else {
		computeTimes.ResponseStatus = resp.StatusCode
	}

	computeTimes.TotalTime = time.Since(start)

	return &computeTimes
}
