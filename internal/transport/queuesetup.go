package transport

import (
	"errors"

	"github.com/rabbitmq/amqp091-go"
)

type MQQueue struct {
	Transport *MQTransport
	QName     string
	DLQName   string
}

type QueueConfig struct {
	QName      string
	DLQName    string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
}
type ChannelQoS struct {
	PrefetchCount int
	PrefetchSize  int
	Global        bool
}

func NewMQQueue(t *MQTransport, c QueueConfig, q ChannelQoS) (*MQQueue, error) {
	_, err := t.ch.QueueDeclare(c.DLQName, c.Durable, c.AutoDelete, c.Exclusive, c.NoWait, nil)
	if err != nil {
		// залогировать
		return nil, errors.New("error during the creating DL queue: " + err.Error())
	}

	args := amqp091.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": c.DLQName,
	}

	_, err = t.ch.QueueDeclare(c.QName, c.Durable, c.AutoDelete, c.Exclusive, c.NoWait, args)
	if err != nil {
		// залогировать
		return nil, errors.New("error during the creating queue: " + err.Error())
	}

	err = t.ch.Qos(q.PrefetchCount, q.PrefetchSize, q.Global)
	if err != nil {
		// залогировать
		return nil, errors.New("error during the setting QoS: " + err.Error())
	}
	return &MQQueue{Transport: t, QName: c.QName, DLQName: c.DLQName}, nil
}

func (q *MQQueue) Consume() (<-chan amqp091.Delivery, error) {
	msg, err := q.Transport.ch.Consume(q.QName, "", false, false, false, false, nil)
	if err != nil {
		// залогировать
		return nil, errors.New("error during consuming message: " + err.Error())
	}
	return msg, nil
}

func (q *MQQueue) Close() error {
	if err := q.Transport.Close(); err != nil {
		return err
	}
	return nil
}
