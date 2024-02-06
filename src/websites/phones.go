package websites

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"mosquitoSwarm/src/db/dao"
	"mosquitoSwarm/src/db/dto"
	"mosquitoSwarm/src/util"
	"mosquitoSwarm/src/websites/web"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// UpdatePhones adds new records to "phones" DB table. Used to scrape for phone nubmers.
func UpdatePhones(baseUrl string, categoryUrls []string, socksProxy string, threadsLimit int) {
	defer util.RecoverAndLog("UpdatePhones")

	log.Info("Phone update triggered")
	var mu = sync.Mutex{}
	concurrencyCh := make(chan struct{}, threadsLimit)

	totalInserted := 0
	for _, categoryUrl := range categoryUrls {
		concurrencyCh <- struct{}{}
		tor := web.OpenAnonymousSession(socksProxy)
		go func(category string) {
			defer func() { <-concurrencyCh }()
			defer util.RecoverAndLog("UpdatePhones : " + category)

			phones := fetchAllPhones(baseUrl, category, tor)

			mu.Lock()
			defer mu.Unlock()
			totalInserted += dao.Dao.SaveNewPhones(phones)
		}(categoryUrl)
	}
	log.Info(fmt.Sprintf("Found %d new phones", totalInserted))
}

func fetchAllPhones(baseUrl string, categoryUrl string, tor *web.AnonymousSession) []dto.Phone {
	var urls []string
	log.Info("Fetching ad list from " + categoryUrl)

	//the first time I actually needed a do/while loop in my life
	for pageNumber, hasNext := 1, true; hasNext; pageNumber++ {
		_, page := tor.GetUrl(categoryUrl + "?page=" + strconv.Itoa(pageNumber))
		adUrls := parseAdList(page)
		if len(adUrls) > 0 {
			urls = append(urls, adUrls...)
		}
		hasNext = hasNextPage(page)
	}

	//send all requests synchronously to avoid getting blocked
	log.Info("Fetching phones from " + categoryUrl)
	var phones []dto.Phone
	for _, url := range urls {
		request := prepareRequestWithPopupBypass(baseUrl + url)
		if request == nil {
			continue
		}
		_, ad := tor.GetRequest(request)
		phone := getPhone(ad)
		if phone.Phone != "" {
			phones = append(phones, phone)
		}
	}
	return removeDuplicates(phones)
}

func parseAdList(htmlPage *goquery.Document) (adLinks []string) {
	adLinks = htmlPage.Find("ul.ads-list-photo").First().Children().
		Filter("li.ads-list-photo-item").
		Filter(":not(.is-adsense):not(.js-booster-inline)").
		Find(".ads-list-photo-item-title").Children().
		Map(func(i int, a *goquery.Selection) string {
			href, _ := a.Attr("href")
			return href
		})
	return adLinks
}

func hasNextPage(htmlPage *goquery.Document) bool {
	return htmlPage.Find("nav.paginator").
		Find(".current").Next().Length() > 0
}

func getPhone(htmlPage *goquery.Document) dto.Phone {
	phone, _ := htmlPage.Find("dl.js-phone-number").
		Find("a").Attr("href")

	if phone != "" && !strings.HasPrefix(phone, "tel:+373") {
		log.Warn("Bad phone prefix in a phone: ", phone)
		return dto.Phone{}
	}
	phone = strings.TrimPrefix(phone, "tel:+373")

	return dto.Phone{Phone: phone}
}

func prepareRequestWithPopupBypass(url string) *http.Request {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithError(err).Error("Failed to create request to get phone!")
		return nil
	}

	request.AddCookie(&http.Cookie{
		Name:    "age_popup_show_guest",
		Value:   "False",
		Expires: time.Now().Add(365 * 24 * time.Hour),
	})
	request.AddCookie(&http.Cookie{
		Name:    "age_popup_show",
		Value:   "False",
		Expires: time.Now().Add(365 * 24 * time.Hour),
	})
	return request
}

func removeDuplicates(phones []dto.Phone) (uniquePhones []dto.Phone) {
MainLoop:
	for _, phone := range phones {
		for _, resultPhone := range uniquePhones {
			if resultPhone.Phone == phone.Phone {
				continue MainLoop
			}
		}
		uniquePhones = append(uniquePhones, phone)
	}
	return
}
