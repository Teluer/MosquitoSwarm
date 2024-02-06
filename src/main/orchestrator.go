package main

import (
	"container/ring"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"mosquitoSwarm/src/admin"
	"mosquitoSwarm/src/config"
	"mosquitoSwarm/src/db/dao"
	"mosquitoSwarm/src/rabbitmq"
	"mosquitoSwarm/src/util"
	"mosquitoSwarm/src/websites"
	"mosquitoSwarm/src/websites/orsen"
	"sync"
	"time"
)

func initializeThings(cfg *config.Config) {
	dao.Dao.OpenDatabaseConnection(cfg.DatabaseDsn)
	dao.Dao.AutoMigrateAll()
	rabbitmq.InitializeRabbitMq(cfg.Rabbit)
	go admin.StartControlPanelServer(&cfg.OrdersRoutineCfg)
}

func scheduleJobs(cfg *config.Config, loc *time.Location) {
	log.Info("Bootstrapping goroutines")

	//wait for the updates to complete, then proceed with orders.
	//this is unnecessary since data integrity checks are in place, keeping this just for lulz
	var wg sync.WaitGroup
	wg.Add(2)
	go websites.UpdateLastNames(&wg, cfg.LastNamesUrl)
	go websites.UpdateFirstNames(&wg, cfg.FirstNamesUrl)
	wg.Wait()

	//using scheduler
	scheduleUpdatePhonesJob(cfg.PhonesCfg, loc, cfg.SocksProxyAddress)

	go manualOrdersJob()
	//everything ready, start sending orders
	go sendOrdersJob(&cfg.OrdersRoutineCfg, loc, cfg.SocksProxyAddress)

	log.Info("Program initialization complete, LET THE FUN BEGIN!")
}

func manualOrdersJob() {
	log.Info("Starting manual orders RabbitMq listener")
	for {
		func() {
			defer util.RecoverAndLog("RabbitMq")
			order := rabbitmq.ConsumeManualOrder()
			orsen.AddManualOrder(order)
			log.Info("Manual order is queued and will be executed by Orders routine")
		}()
	}
}

func scheduleUpdatePhonesJob(cfg config.PhonesConfig, loc *time.Location, socksProxy string) {
	startTime := time.Now().Add(cfg.UpdatePhonesStartDelay)

	s := gocron.NewScheduler(loc).Every(cfg.UpdatePhonesInterval).StartAt(startTime)
	_, err := s.Do(func() {
		websites.UpdatePhones(cfg.PhonesBaseUrl, cfg.PhoneUrls, socksProxy, cfg.UpdatePhonesThreadsLimit)
	})

	if err != nil {
		log.WithError(err).Error("Failed to start UpdatePhones job")
	} else {
		log.Info("Starting UpdatePhones job")
		s.StartAsync()
	}
}

func sendOrdersJob(cfg *config.OrdersRoutineConfig, loc *time.Location, socksProxy string) {
	log.Info("Starting send orders routine")

	if cfg.StartDelay {
		log.Info("Configured to delay the first order")
		sleepDuration := time.Duration(float64(cfg.SendOrdersMaxInterval) * rand.Float64())
		log.Infof("sendOrdersJob: scheduling order in %.2f minutes", float64(sleepDuration)/float64(time.Minute))
		time.Sleep(sleepDuration)
	} else {
		log.Warn("Sending initial order without delay!")
	}

	//populate thread ids to run selenium on different ports
	routineIds := ring.New(cfg.SendOrdersMaxThreads)
	for i := 0; i < cfg.SendOrdersMaxThreads; i++ {
		routineIds.Value = i
		routineIds = routineIds.Next()
	}

	concurrencyCh := make(chan struct{}, cfg.SendOrdersMaxThreads)

	for {
		concurrencyCh <- struct{}{}
		log.Info("sendOrdersJob: OrderSender flow triggered")
		sleepAtNight(cfg, loc)

		jobStart := time.Now()

		readyToGo := dao.Dao.ValidateDataIntegrity()
		ordersEnabled := cfg.SendOrdersEnabled

		if readyToGo && ordersEnabled {
			order := orsen.NewOrderSender(&cfg.OrdersCfg, socksProxy, routineIds.Value.(int), concurrencyCh)
			routineIds = routineIds.Next()

			go order.OrderItem()
		} else {
			if !readyToGo {
				log.Warn("Cannot send orders due to empty database tables, please check DB!")
			}
			if !ordersEnabled {
				log.Info("SendOrdersEnabled = false, not sending anything")
			}
		}
		jobDuration := time.Now().Sub(jobStart)
		sleepDuration := time.Duration(float64(cfg.SendOrdersMaxInterval)*rand.Float64()) - jobDuration
		log.Infof("sendOrdersJob: scheduling next order in %.2f minutes", float64(sleepDuration)/float64(time.Minute))
		time.Sleep(sleepDuration)
	}
}

func sleepAtNight(cfg *config.OrdersRoutineConfig, loc *time.Location) {
	year, month, day := time.Now().In(loc).Date()
	midnight := time.Date(year, month, day, 0, 0, 0, 0, loc)

	currentTime := time.Now().In(loc)
	startTime := midnight.Add(cfg.DayStart)
	endTime := midnight.Add(cfg.DayEnd)

	var sleepDuration time.Duration
	if currentTime.Before(startTime) {
		sleepDuration = startTime.Sub(currentTime)
	} else if currentTime.After(endTime) {
		sleepDuration = startTime.Add(time.Hour * 24).Sub(currentTime)
	} else {
		return
	}
	log.Info("sendOrdersJob: Beyond work hours, sleeping until " +
		time.Now().Add(sleepDuration).In(loc).Format("2006-01-02 15:04:05"))
	time.Sleep(sleepDuration)
}
