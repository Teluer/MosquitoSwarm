# Mosquito Swarm Project #

## Backstory ##
- A fraudulent local online shop sold a cheaper and worse product than the one in the description.
- The shop refused to return the product in spite of their 14-day return policy.
- A brief investigation showed that the shop has no captcha, but it does have a fast order form that requests a call back from the operator.

## Project purpose ##
The aim is to study and practice GoLang and to deliver the mild harassment this shop deserves for mistreating its customers.

The project submits lots of random orders to make the operator call people all day long and ask if they ordered something.

The chosen phone base consists of escort workers, because they are often busy and therefore each order will likely produce several call back attempts from the operator.

## Used technologies ##
- Golang (with gorm, gocron, goquery, logrus, testify etc)
- MySql database
- RabbitMq (not really necessary, included just for practice)
- TOR
- Selenium
- Docker

## Project description ##
- Run indefinitely
- Scrape an escort website for phone numbers
- Gather and combine popular moldovan first and last names from the Internet
- Submit random fast orders in the online shop
- Generated orders should be unique and indistinguishable from real orders (see [Order Concealing](#order-concealing) section)
- Send orders using TOR to make it impossible to block the server IP or trace the source
- Configurable order rates and schedule
- Admin panel to change some configs on the fly
- Keep track of all submitted orders in DB
- Keep logs

## Order Concealing ##
- Send orders at random intervals
- Keep mean order rate low enough (not less than 30-60 minutes) to avoid causing panic
- Send orders from random IP addresses close to Moldova
- Use random user-agent header
- Use random phone and name formats
- Avoid reusing phones and names unless there are no new values available
- Gather cookies from the server responses and attach them to the subsequent requests
- Visit order success page
- For selenium flow, make random user actions while filling the order form (e.g. move mouse around) to deceive anti-bot scripts

## Project story ##
- The first working MVP instance was built within a week.
- Order #45: the web shop enabled Google ReCaptcha. Ability to bypass captcha was partially implemented but couldn't be tested, because the webshop disabled captcha a day later for unknown reasons.
- Order #165: a control order with a friend's number was submitted to check if the operator still calls, and it worked.
- Order #235: the website enabled another (surprisingly easy) captcha. Selenium flow and captcha-solving logic were implemented a day later.
- Order #249: the website removed captcha again. The program didn't need any changes and continued working normally. Order interval was briefly reduced to 4 minutes 30 seconds as a punishment.
- Order #362: running out of phones, adding new phone categories (private massage therapists and some others).
- Order #366: Google ReCaptcha came back. And it targets Tor IP addresses. Further development would require a paid VPN and a paid captcha solver extension. Solution: treat strong captcha as a sign of project success and wrap things up.
- Parallel order sending implemented. The flow successfully reaches recaptcha challenge (and then fails).
- After a Docker update, RabbitMq fails to start due to a missing library. Since the project is not functional anyway, we call it a day.
