package web

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
	"io"
	"net/http"
)

type SocksConnection struct {
	client *http.Client
}

func connectToSocksProxy(proxyUrl string) *SocksConnection {
	log.Debug("Opening TOR session")
	dialer, err := proxy.SOCKS5("tcp", proxyUrl, nil, proxy.Direct)
	if err != nil {
		log.WithError(err).Error("Failed to connect to SOCKS proxy!")
		panic(err)
	}

	var ts SocksConnection
	ts.client = &http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}
	log.Debug("Established SOCKS connection successfully")
	return &ts
}

func (ts *SocksConnection) socksRequest(req *http.Request) (*http.Response, []byte) {
	if ts.client == nil {
		log.Error("SOCKS client not initialized")
		panic("need to init SOCKS client first!")
	}

	log.Debug("Sending anonymous request to url:" + req.URL.String())
	resp, err := ts.client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("failed to send anonymous request %v", err))
	}
	defer resp.Body.Close()
	log.Debug("Anonymous request returned status " + resp.Status)

	body, _ := io.ReadAll(resp.Body)
	return resp, body
}

func (ts *SocksConnection) socksGet(url string) (*http.Response, []byte) {
	if ts.client == nil {
		log.Error("SOCKS client not initialized")
		panic("need to init SOCKS client first!")
	}

	log.Debug("Sending anonymous GET to url:" + url)
	resp, err := ts.client.Get(url)
	if err != nil {
		log.WithError(err).Error("Anonymous GET request failed")
		panic(err)
	}
	defer resp.Body.Close()

	log.Debug("Got status", resp.Status, " to anonymous GET to url:", url)

	body, _ := io.ReadAll(resp.Body)
	return resp, body
}
