package web

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type AnonymousSession struct {
	session *SocksConnection
}

func GetUnsafe(url string) *goquery.Document {
	resp, err := http.Get(url)
	if err != nil {
		log.WithError(err).Error("Unsafe GET request failed")
		panic(err)
	}
	defer resp.Body.Close()
	log.Info("Got status", resp.Status, " to unsafe GET to url:", url)

	body, _ := io.ReadAll(resp.Body)
	return extractDocumentFromResponseBody(body)
}

func RequestUnsafe(req *http.Request) *goquery.Document {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithError(err).Error("Unsafe request failed")
		panic(err)
	}
	defer resp.Body.Close()
	log.Debug("Got status", resp.Status, " to unsafe request to url:", req.URL.String())

	body, _ := io.ReadAll(resp.Body)
	return extractDocumentFromResponseBody(body)
}

func OpenAnonymousSession(proxy string) *AnonymousSession {
	return &AnonymousSession{
		session: connectToSocksProxy(proxy),
	}
}

func (as *AnonymousSession) SendRequest(req *http.Request) (response *http.Response, body []byte) {
	return as.session.socksRequest(req)
}

func (as *AnonymousSession) GetUrl(url string) (*http.Response, *goquery.Document) {
	resp, responseBody := as.session.socksGet(url)
	return resp, extractDocumentFromResponseBody(responseBody)
}

func (as *AnonymousSession) GetRequest(req *http.Request) (*http.Response, *goquery.Document) {
	resp, body := as.SendRequest(req)
	return resp, extractDocumentFromResponseBody(body)
}

func (as *AnonymousSession) FetchCookies(url string) []*http.Cookie {
	var resp *http.Response
	resp, _ = as.session.socksGet(url)
	return resp.Cookies()
}

func extractDocumentFromResponseBody(body []byte) *goquery.Document {
	reader := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.WithError(err).Error("Failed to parse response body")
		panic(fmt.Sprintf("while parsing response body: %v", err))
	}

	return doc
}
