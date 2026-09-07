package editor

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// Listen opens a TCP listener that accepts loopback addresses only. An empty
// host means 127.0.0.1.
func Listen(addr string) (net.Listener, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("listen address %q: %w", addr, err)
	}
	if host == "" {
		host = "127.0.0.1"
	}
	if host != "localhost" {
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			return nil, fmt.Errorf("the editor listens on loopback only; refusing %q", addr)
		}
	}
	return net.Listen("tcp", net.JoinHostPort(host, port))
}

// LoopbackOnly refuses any request whose Host is not loopback, so a
// DNS-rebinding page cannot reach the editor through a browser, and passes
// the rest to next.
func LoopbackOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !loopbackHost(r.Host) {
			http.Error(w, fmt.Sprintf("host %q is not loopback", r.Host), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// loopbackHost reports whether a Host header names 127.0.0.1, [::1], or
// localhost, with or without a port.
func loopbackHost(host string) bool {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}
	return host == "127.0.0.1" || host == "::1" || strings.EqualFold(host, "localhost")
}
