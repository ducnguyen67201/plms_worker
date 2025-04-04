package AppConfig

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type MQClient struct {
	Conn    *amqp091.Connection
	Channel *amqp091.Channel
}

var MQ *MQClient

func InitRabbitMQ(url string) (*MQClient, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	MQ = &MQClient{
		Conn:    conn,
		Channel: ch,
	}

	fmt.Print("RabbitMQ connected successfully\n")

	return MQ, nil
}
