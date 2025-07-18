package transport

import (
	"errors"

	"github.com/rabbitmq/amqp091-go"
)

type MQTransport struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
}

// Don't forget to call MQTransport.Close()
func NewMQTransport(dsn string) (*MQTransport, error) {
	conn, err := amqp091.Dial(dsn)
	if err != nil {
		// залогировать
		return nil, errors.New("error during the connecting: " + err.Error())
	}
	ch, err := conn.Channel()
	if err != nil {
		// залогировать
		return nil, errors.New("error during the creating channel: " + err.Error())
	}
	return &MQTransport{conn: conn, ch: ch}, nil
}

func (mq *MQTransport) Close() error {
	var errs []error
	if mq.ch != nil {
		if err := mq.ch.Close(); err != nil {
			errs = append(errs, err)
		}
		mq.ch = nil
	}

	if mq.conn != nil {
		if err := mq.conn.Close(); err != nil {
			errs = append(errs, err)
		}
		mq.conn = nil
	}
	return errors.Join(errs...)
}
