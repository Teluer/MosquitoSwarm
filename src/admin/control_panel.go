package admin

import (
	log "github.com/sirupsen/logrus"
	"html/template"
	"mosquitoSwarm/src/config"
	"mosquitoSwarm/src/db/dao"
	"mosquitoSwarm/src/rabbitmq"
	"mosquitoSwarm/src/websites/orsen"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var mu sync.Mutex

func StartControlPanelServer(cfg *config.OrdersRoutineConfig) {
	log.Infof("Starting Control Panel server at: %s", "localhost:8008/admin")

	http.HandleFunc("/admin", controlPanelHandler(cfg)) // each request calls handler
	http.HandleFunc("/conf", configHandler(cfg))
	http.HandleFunc("/order", orderHandler)
	http.HandleFunc("/customer", customerHandler(cfg.OrdersCfg.PhonePrefixes))

	log.Error(http.ListenAndServe("0.0.0.0:8008", nil))
}

func controlPanelHandler(cfg *config.OrdersRoutineConfig) func(http.ResponseWriter, *http.Request) {
	if bytes, err := os.ReadFile("controlPanel.html"); err == nil {
		templ := string(bytes)

		return func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			log.Info("Accessing control panel from: ", r.RemoteAddr)

			data := &struct {
				OrdersInterval float64
				OrdersEnabled  bool
			}{
				OrdersInterval: float64(cfg.SendOrdersMaxInterval.Nanoseconds()) / float64(time.Minute),
				OrdersEnabled:  cfg.SendOrdersEnabled,
			}

			t, err := template.New("form").Parse(templ)
			if err != nil {
				log.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = t.Execute(w, data)
			if err != nil {
				log.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			mu.Unlock()
		}
	} else {
		log.Error("couldn't read controlPanel.html: ", err)
		panic("couldn't read controlPanel.html!")
	}
}

func configHandler(cfg *config.OrdersRoutineConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//imagine error handling here
		frequencyMinutes, _ := strconv.ParseFloat(r.FormValue("frequency"), 64)
		ordersEnabled := r.FormValue("shouldSend") == "on"

		cfg.SendOrdersMaxInterval = time.Duration(float64(time.Minute) * frequencyMinutes)
		cfg.SendOrdersEnabled = ordersEnabled

		log.Infof("Updated configs: interval=%v, sending=%v", cfg.SendOrdersMaxInterval, cfg.SendOrdersEnabled)
		message := "Configs updated. changes will take effect after the next scheduled order is sent!"
		w.Write([]byte(message))
	}
}

func orderHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	phone := r.FormValue("phone")

	if name == "" && phone == "" {
		w.Write([]byte("Both fields cannot be blank!"))
		return
	}

	order := rabbitmq.ManualOrder{
		Name:  name,
		Phone: phone,
	}

	err := rabbitmq.PublishManualOrder(&order)

	if err != nil {
		w.Write([]byte("Manual order submission failed! See logs for more info."))
	} else {
		w.Write([]byte("Manual order submitted. Will be sent at scheduled time."))
	}
}

func customerHandler(phonePrefixes string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name, phone := orsen.CreateRandomCustomer(dao.Dao, phonePrefixes)
		w.Write([]byte(name + ", " + phone))
	}
}
