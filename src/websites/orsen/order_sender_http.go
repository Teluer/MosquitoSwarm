package orsen

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"mosquitoSwarm/src/util"
	"mosquitoSwarm/src/websites/web"
	"net/http"
	"net/url"
	"strings"
)

type OrderSuccess struct {
	Success  int    `json:"success"`
	Redirect string `json:"redirect_location"`
}

func (ord *OrderSender) orderItemWithCustomerHttp(name, phone, itemId, link string, tor *web.AnonymousSession) {
	log.Info(fmt.Sprintf("Sending order for (%s, %s, %s) with cookies from %s to %s ",
		name, phone, itemId, link, ord.cfg.TargetOrderLink))

	cookies := tor.FetchCookies(link)
	log.Info("Got cookies: ", cookies)

	req := ord.prepareOrderRequest(ord.cfg.TargetOrderLink, name, phone, itemId, link, cookies)
	resp, body := tor.SendRequest(req)
	log.Info(string(body))
	cookies = append(cookies, resp.Cookies()...)

	var responseBody OrderSuccess
	json.Unmarshal(body, &responseBody)

	if responseBody.Success != 1 {
		log.Error("Response body is not success=1!")
		return
	}

	log.Info("Visiting order page to reproduce user behaviour")
	req = ord.prepareOrderSuccessGetRequest(responseBody.Redirect, link, cookies)
	resp, _ = tor.SendRequest(req)
	cookies = append(cookies, resp.Cookies()...)

	log.Info("Confirming payment method for the new order " +
		"(NOT, because the website disabled this button)")
	//req = prepareConfirmOrderRequest(responseBody.Redirect, cookies)
	//resp, body = tor.SendRequest(req)

	log.Info("Sent order successfully")
}

func (ord *OrderSender) findRandomItem(tor *web.AnonymousSession) (id string, link string) {
	caregoryIndex := rand.Intn(len(ord.cfg.TargetCategories))
	randomCategory := ord.cfg.TargetCategories[caregoryIndex]

	log.Info("Fetching random item from category " + randomCategory)

	_, page := tor.GetUrl(randomCategory)

	items := page.Find("a.product_preview__name_link")
	randomItem := rand.Intn(items.Length())
	items.EachWithBreak(func(i int, item *goquery.Selection) bool {
		if i == randomItem {
			id, _ = item.Attr("data-product")
			link, _ = item.Attr("href")
			return false
		}
		return true
	})

	link = ord.cfg.TargetBaselink + link
	log.Info("Will order the following item: " + id + " " + link)
	return id, link
}

func (ord *OrderSender) prepareOrderRequest(target, name, phone, itemId, referer string, cookies []*http.Cookie) *http.Request {
	//create request body
	formData := url.Values{}
	formData.Set("variant_id", itemId)
	formData.Set("amount", "")
	formData.Set("IsFastOrder", "true")
	formData.Set("name", name)
	formData.Set("phone", phone)
	formData.Set("delivery_id", "4")

	// Encode the form data
	body := strings.NewReader(formData.Encode())

	//create request
	request, err := http.NewRequest("POST", target, body)
	if err != nil {
		log.WithError(err).Error("Failed to create request for order!")
		panic("failed to create request!")
	}

	//set cookies previously returned by the server
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	//set proper headers
	request.Header.Set("User-Agent", util.RandomUserAgent())
	request.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	request.Header.Set("Accept-Language", "en-US,en;q=0.5")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	request.Header.Set("Origin", "**REMOVED**")
	request.Header.Set("DNT", "1")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Referer", referer)
	request.Header.Set("Sec-Fetch-Dest", "empty")
	request.Header.Set("Sec-Fetch-Mode", "cors")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	request.Header.Set("Host", "**REMOVED**")

	return request
}

func (ord *OrderSender) prepareOrderSuccessGetRequest(target, referer string, cookies []*http.Cookie) *http.Request {
	request, err := http.NewRequest("GET", target, nil)
	if err != nil {
		log.WithError(err).Error("Failed to create request for order!")
		panic("failed to create request!")
	}

	//set cookies previously returned by the server
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	request.Header.Set("User-Agent", util.RandomUserAgent())
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	request.Header.Set("Accept-Language", "en-US,en;q=0.5")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	request.Header.Set("Origin", "**REMOVED**")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Referer", referer)
	request.Header.Set("Sec-Fetch-Dest", "document")
	request.Header.Set("Upgrade-Insecure-Requests", "1")
	request.Header.Set("Sec-Fetch-Mode", "navigate")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	request.Header.Set("Sec-Fetch-User", "?1")

	return request
}

func (ord *OrderSender) prepareConfirmOrderRequest(target string, cookies []*http.Cookie) *http.Request {
	formData := url.Values{}
	formData.Set("payment_method_id", "23")
	formData.Set("checkout", "Применить")
	// Encode the form data
	body := strings.NewReader(formData.Encode())

	request, err := http.NewRequest("POST", target, body)
	if err != nil {
		log.WithError(err).Error("Failed to create request for order!")
		panic("failed to create request!")
	}

	//set cookies previously returned by the server
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	request.Header.Set("User-Agent", util.RandomUserAgent())
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	request.Header.Set("Accept-Language", "en-US,en;q=0.5")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	request.Header.Set("Origin", "**REMOVED**")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Referer", target)
	request.Header.Set("Sec-Fetch-Dest", "document")
	request.Header.Set("Upgrade-Insecure-Requests", "1")
	request.Header.Set("Sec-Fetch-Mode", "navigate")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	request.Header.Set("Sec-Fetch-User", "?1")

	return request
}
