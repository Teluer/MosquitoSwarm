package config

import (
	"time"
)

// Config contains all the properties including embedded property sets.
type Config struct {
	TimeZone string `properties:"timezone"`

	OrdersRoutineCfg OrdersRoutineConfig `properties:"ordersroutine"`
	Rabbit           RabbitConfig        `properties:"rabbit"`

	SocksProxyAddress string `properties:"socks.proxy"`
	FirstNamesUrl     string `properties:"url.firstnames"`
	LastNamesUrl      string `properties:"url.lastnames"`
	DatabaseDsn       string `properties:"db.dsn"`

	PhonesCfg PhonesConfig `properties:"phones"`
}

type OrdersRoutineConfig struct {
	SendOrdersMaxInterval time.Duration `properties:"send.interval.max"`
	SendOrdersMaxThreads  int           `properties:"send.threads.limit"`
	StartDelay            bool          `properties:"start.delay"`
	DayStart              time.Duration `properties:"day.start"`
	DayEnd                time.Duration `properties:"day.end"`
	SendOrdersEnabled     bool          `properties:"orders.enabled""`
	OrdersCfg             OrdersConfig  `properties:"orders"`
}

type OrdersConfig struct {
	SeleniumEnabled bool `properties:"selenium.enabled"`
	//loading string array as string to avoid space trimming by the properties lib
	PhonePrefixes    string   `properties:"phone.prefixes"`
	TargetBaselink   string   `properties:"target.base"`
	TargetOrderLink  string   `properties:"target.order"`
	TargetCategories []string `properties:"target.categories"`
}

type RabbitConfig struct {
	Host      string `properties:"host"`
	QueueName string `properties:"queue"`
}

type PhonesConfig struct {
	UpdatePhonesInterval     time.Duration `properties:"update.interval"`
	UpdatePhonesThreadsLimit int           `properties:"update.threads.limit"`
	UpdatePhonesStartDelay   time.Duration `properties:"start.delay"`
	PhonesBaseUrl            string        `properties:"base"`
	PhoneUrls                []string      `properties:"categories"`
}
