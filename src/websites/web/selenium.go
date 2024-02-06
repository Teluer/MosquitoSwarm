package web

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"math/rand"
	"mosquitoSwarm/src/util"
	"strconv"
	"strings"
	"time"
)

type Selenium struct {
	service *selenium.Service
	driver  *selenium.WebDriver
}

// OpenChromeSession opens a new Chrome instance and navigates to the provided url
func OpenChromeSession(url, socksProxy string, routineId int) *Selenium {
	port := 4444 + routineId

	userAgent := util.RandomUserAgent()
	log.Info("Starting selenium with user-agent = ", userAgent)

	service, err := selenium.NewChromeDriverService("chromedriver", port)
	if err != nil {
		log.Errorf("Failed to start the WebDriver service: %v", err)
		panic(err)
	}

	caps := selenium.Capabilities{
		"browserName": "chrome",
	}
	caps.AddProxy(selenium.Proxy{
		Type:         selenium.Manual,
		SOCKS:        socksProxy,
		SOCKSVersion: 5,
	})
	cc := chrome.Capabilities{
		W3C: false,
		Args: []string{
			"--no-sandbox",
			"--headless", // Run Chrome in headless mode (without UI)
			"--user-agent=" + userAgent,
			"--remote-debugging-port=" + strconv.Itoa(9222+routineId),
			"--temp-profile=tmpp" + strconv.Itoa(routineId),
		},
		Prefs: map[string]interface{}{
			"enable_do_not_track": true,
		},
	}
	caps.AddChrome(cc)

	// Start a Tor browser session
	webDriver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		log.Errorf("Failed to start browser session: %v", err)
		panic(err)
	}
	webDriver.SetPageLoadTimeout(time.Minute * 3)
	webDriver.SetImplicitWaitTimeout(time.Minute * 3)

	err = webDriver.Get(url)
	if err != nil {
		log.Errorf("Failed to navigate to url: %v", err)
		panic(err)
	}

	return &Selenium{service: service, driver: &webDriver}
}

func (s *Selenium) Close() {
	s.service.Stop()
	(*s.driver).Quit()
}

func (s *Selenium) GetDocument() *goquery.Document {
	webDriver := *s.driver
	html, _ := webDriver.PageSource()
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	return doc
}

func (s *Selenium) Click(selector string) {
	webDriver := *s.driver
	element, err := webDriver.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		log.Errorf("Failed to find element with selector: %q, %v", selector, err)
		panic(err)
	}

	err = element.Click()
	if err != nil {
		log.Errorf("Failed to click element: %v", err)
		panic(err)
	}
}

func (s *Selenium) WaitForRedirect(url string) {
	webDriver := *s.driver

	condition := func(wd selenium.WebDriver) (bool, error) {
		currentUrl, err := wd.CurrentURL()
		if err != nil {
			return false, err
		}
		return strings.Contains(currentUrl, url), nil
	}

	err := webDriver.Wait(condition)
	if err != nil {
		log.Errorf("Oder probably failed! Wait for redirect failed: %v", err)
		panic(err)
	}
}

func (s *Selenium) Input(selector, value string) {
	webDriver := *s.driver
	element, err := webDriver.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		log.Errorf("Failed to find element with selector: %q, %v", selector, err)
		panic(err)
	}

	err = element.SendKeys(value)
	if err != nil {
		log.Errorf("Failed to input: %v", err)
		panic(err)
	}
}

func (s *Selenium) Idle(minSeconds, maxSeconds float64) {
	seconds := (maxSeconds-minSeconds)*rand.Float64() + minSeconds
	sleepDuration := time.Duration(float64(time.Second) * seconds)
	time.Sleep(sleepDuration)
}

// MoveAround imitates a user, performing random mouse movements for some time
// atLeast - minimal amount of time to move around.
func (s *Selenium) MoveAround(atLeast time.Duration) {
	webDriver := *s.driver
	body, err := webDriver.FindElement(selenium.ByTagName, "body")
	if err != nil {
		log.Error("Couldn't find body, won't move around")
		return
	}
	size, err := body.Size()
	if err != nil {
		log.Error("Couldn't get body size, won't move around")
		return
	}

	shouldEnd := time.Now().Add(atLeast)
	for time.Now().Before(shouldEnd) {
		x := rand.Intn(size.Width) - size.Width/2   //left or right movement
		y := rand.Intn(size.Height) - size.Height/2 //up or down movement
		err = body.MoveTo(x, y)
		if err != nil {
			log.Error("Failed to move mouse")
		}
	}
}

func (s *Selenium) EnterCaptcha(value string) {
	webDriver := *s.driver
	element, err := webDriver.FindElement(selenium.ByCSSSelector, "#fn_fast_order input.form__input_captcha")
	if err != nil {
		log.Errorf("Failed to find captcha field: %v", err)
		panic(err)
	}

	err = element.SendKeys(value)
	if err != nil {
		log.Errorf("Failed to enter captcha: %v", err)
		panic(err)
	}
}

// Deprecated: this fails with socks proxy as ReCaptcha detects TOR exit nodes.
func (s *Selenium) SolveReCaptcha() {
	webDriver := *s.driver

	time.Sleep(time.Second * 3)
	challenge, err := webDriver.FindElement(selenium.ByCSSSelector, "iframe[title='recaptcha challenge expires in two minutes']")
	if err != nil {
		log.Warnf("Failed to find captcha challenge frame: %v", err)
		return
	}
	webDriver.SwitchFrame(challenge)

	button, err := webDriver.FindElement(selenium.ByCSSSelector, "div.help-button-holder")
	if err != nil {
		log.Errorf("Failed to find captcha-solving button: %v", err)
		panic(err)
	}

	button.Click()
	webDriver.SwitchFrame(nil)

	//the rest of this method was removed as it doesn't apply to the website captcha implementation anymore.
}
