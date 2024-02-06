package orsen

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mosquitoSwarm/src/websites/orsen/captcha"
	"mosquitoSwarm/src/websites/web"
)

func (ord *OrderSender) orderItemWithCustomerSelenium(name, phone, itemId, link string) {
	//simpler to do with just tor
	log.Info(fmt.Sprintf("Sending order for (%s, %s, %s) via Selenium",
		name, phone, itemId))

	selenium := web.OpenChromeSession(link, ord.socksProxy, ord.threadId)
	defer selenium.Close()

	selenium.MoveAround(2)
	selenium.Click("a.fn_fast_order_button")
	selenium.MoveAround(3)
	selenium.Input("#fn_fast_order input.fn_validate_fast_name", name)
	selenium.MoveAround(2)
	selenium.Input("#fn_fast_order input.fn_validate_fast_phone", phone)
	selenium.MoveAround(2)
	ord.solveYandexCaptcha(selenium)
	selenium.MoveAround(2)
	selenium.Click("#fn_fast_order input.fn_fast_order_submit")
	//selenium.SolveReCaptcha()
	selenium.WaitForRedirect("/order/")

	log.Info("Sent order successfully")
}

func (ord *OrderSender) solveYandexCaptcha(selenium *web.Selenium) {
	doc := selenium.GetDocument()
	challenge := doc.Find("#fn_fast_order").Find("div.secret_number").Text()
	if challenge == "" {
		log.Info("No arithmetic captcha found, skipping captcha answer")
		return
	}

	log.Info("Arithmetic captcha found, solving: " + challenge)
	answer := captcha.SolveArithmeticCaptcha(challenge)
	log.Info("Captcha solution: ", answer)

	selenium.EnterCaptcha(answer)
}
