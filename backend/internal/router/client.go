package router

import (
	"net"
	"net/http"
	"time"
)

var proxyClient *http.Client

func NewProxyClient() *http.Client {
	if proxyClient != nil {
		return proxyClient
	}

	proxyClient = &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ForceAttemptHTTP2:     true,
			MaxIdleConnsPerHost:   10,
		},
	}

	return proxyClient
}
