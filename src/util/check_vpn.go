package util

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// PanicIfVpnNotEnabled geolocates current IP address and panics if current location is Moldova.
func PanicIfVpnNotEnabled() {
	resp, err := http.Get("https://api.ipgeolocation.io/ipgeo?apiKey=e202c70dc7b04e0b83d69b27d2c16997")
	if err != nil {
		log.WithError(err).Error("Failed to geolocate")
		panic(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var country struct {
		Country string `json:"country_code2"`
	}
	err = json.Unmarshal(body, &country)
	if err != nil || country.Country == "MD" {
		panic("VPN not enabled!")
	}
}
