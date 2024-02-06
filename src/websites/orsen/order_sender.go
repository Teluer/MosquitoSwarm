package orsen

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mosquitoSwarm/src/config"
	"mosquitoSwarm/src/db/dao"
	"mosquitoSwarm/src/db/dto"
	"mosquitoSwarm/src/rabbitmq"
	"mosquitoSwarm/src/util"
	"mosquitoSwarm/src/websites/web"
	"time"
)

var manualOrders []*rabbitmq.ManualOrder

func AddManualOrder(order *rabbitmq.ManualOrder) {
	manualOrders = append(manualOrders, order)
}

type OrderSender struct {
	cfg                *config.OrdersConfig
	socksProxy         string
	threadId           int
	concurrencyCh      chan struct{}
	currentTransaction dao.Database
}

func NewOrderSender(cfg *config.OrdersConfig, socksProxy string, threadId int, concurrencyCh chan struct{}) *OrderSender {
	return &OrderSender{
		cfg:           cfg,
		socksProxy:    socksProxy,
		threadId:      threadId,
		concurrencyCh: concurrencyCh,
	}
}

func (ord *OrderSender) OrderItem() {
	//notify channel when order is sent
	defer func() { <-ord.concurrencyCh }()

	ord.currentTransaction = dao.Dao.OpenTransaction()
	defer util.RecoverAndLogAndDo("Orders", ord.currentTransaction.RollbackTransaction)

	//check manually prepared orders, if there are no manual orders then make random order
	if !ord.executeManualOrder() {
		log.Info("Sending random order")
		name, phone := CreateRandomCustomer(ord.currentTransaction, ord.cfg.PhonePrefixes)
		ord.orderItemWithCustomer(name, phone)
	}
	ord.currentTransaction.CommitTransaction()
}

func (ord *OrderSender) executeManualOrder() bool {
	log.Info("Checking if should send manual orders")
	if len(manualOrders) == 0 {
		log.Info("Manual orders not found")
		return false
	}

	order := manualOrders[0]
	if order.Name == "" {
		order.Name = generateName(ord.currentTransaction)
	}
	if order.Phone == "" {
		order.Phone = generatePhone(ord.currentTransaction, ord.cfg.PhonePrefixes)
	}

	log.Info(fmt.Sprintf("Sending manual order for %s %s", order.Name, order.Phone))
	ord.orderItemWithCustomer(order.Name, order.Phone)
	manualOrders = manualOrders[1:]
	return true
}

func (ord *OrderSender) orderItemWithCustomer(name, phone string) {
	tor := web.OpenAnonymousSession(ord.socksProxy)
	itemId, link := ord.findRandomItem(tor)

	if ord.cfg.SeleniumEnabled {
		ord.orderItemWithCustomerSelenium(name, phone, itemId, link)
	} else {
		ord.orderItemWithCustomerHttp(name, phone, itemId, link, tor)
	}
	ord.saveOrderHistory(name, phone, itemId)
}

func (ord *OrderSender) saveOrderHistory(name, phone, itemId string) {
	var record = dto.OrderHistory{
		Phone:         phone,
		Name:          name,
		ItemId:        itemId,
		OrderDateTime: time.Now(),
	}

	ord.currentTransaction.Insert(&record)
}
