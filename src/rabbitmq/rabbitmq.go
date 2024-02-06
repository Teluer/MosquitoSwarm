package rabbitmq

import (
	"encoding/json"
	"github.com/silenceper/pool"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"mosquitoSwarm/src/config"
	"mosquitoSwarm/src/util"
	"time"
)

type ManualOrder struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

var rabbit struct {
	rabbitPool *pool.Pool
	queue      string
}

// InitializeRabbitMq tests connection to RabbitMq and declares a single queue for manual orders flow.
// cfg should contain a valid rabbitmq host and the desired queue name.
func InitializeRabbitMq(cfg config.RabbitConfig) {
	log.Info("Initializing rabbitmq connection")
	//wait for rabbitmq to initialize
	TestConnection(cfg.Host)

	rabbit.rabbitPool = openConnectionPool(cfg)
	rabbit.queue = cfg.QueueName

	// Declare queue
	ch := channel()
	defer ch.Close()

	_, err := ch.QueueDeclare(
		cfg.QueueName, // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Errorf("Failed to declare manual orders queue: %v", err)
	}
}

// TestConnection tries to connect to the MQ.
// It returns true if connection could be established and false otherwise.
func TestConnection(url string) bool {
	const maxRetries = 5
	const retryTimeout = time.Second * 5

	for i := 1; i <= maxRetries; i++ {
		log.Infof("Connecting to RabbitMQ (attempt %d/%d)...", i, maxRetries)
		conn, err := amqp.Dial(url)
		if err == nil {
			conn.Close()
			return true
		}

		log.Warnf("Failed to connect to RabbitMQ: %v", err)
		time.Sleep(retryTimeout)
	}
	log.Errorf("Failed to connect to connect to RabbitMQ after %d attempts!", maxRetries)
	return false
}

func ConsumeManualOrder() *ManualOrder {
	defer util.RecoverAndLog("RabbitMq")

	ch := channel()
	defer ch.Close()
	messages, err := ch.Consume(
		rabbit.queue, // queue
		"orders",     // consumer
		true,         // auto-acknowledge messages
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Errorf("Failed to register a consumer: %v", err)
		panic(err)
	}

	message := <-messages
	log.Infof("Received message: %s", string(message.Body))

	var manualOrder ManualOrder
	err = json.Unmarshal(message.Body, &manualOrder)
	if err != nil {
		log.Errorf("Failed to unmarshal message: %v", err)
		// Handle unmarshal error
		return nil
	}
	return &manualOrder
}

func PublishManualOrder(order *ManualOrder) error {
	ch := channel()
	defer ch.Close()

	// Publish a message to the queue
	message, err := json.Marshal(*order)
	if err != nil {
		log.Error("Failed to marshal manual order!")
		return err
	}

	err = ch.Publish(
		"",
		rabbit.queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		})
	if err != nil {
		log.Errorf("Failed to publish a message: %v", err)
		return err
	} else {
		log.Infof("Message sent: %s", message)
	}
	return nil
}

func openConnectionPool(cfg config.RabbitConfig) *pool.Pool {
	poolConfig := &pool.Config{
		InitialCap: 1,
		MaxCap:     10,
		MaxIdle:    5,
		Factory: func() (interface{}, error) {
			return amqp.Dial(cfg.Host)
		},
		Close: func(v interface{}) error {
			conn := v.(amqp.Connection)
			return conn.Close()
		},
	}
	connPool, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ connection pool: %v", err)
	}
	return &connPool
}

func channel() *amqp.Channel {
	conn, err := (*rabbit.rabbitPool).Get()
	if err != nil {
		log.Panic("Failed to acquire connection from pool:", err)
		panic(err)
	}

	// Convert the connection to the appropriate type (amqp.Connection)
	rabbitMQConn := conn.(*amqp.Connection)

	ch, err := rabbitMQConn.Channel()
	if err != nil {
		log.Panic("Failed to open a channel:", err)
		panic(err)
	}
	return ch
}
