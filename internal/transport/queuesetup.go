package transport

import (
	"errors"

	"github.com/rabbitmq/amqp091-go"
)

// КОРОЧЕ ПЕРЕДЕЛАТЬ С ДВУМЯ ОЧЕРЕДЯМИ, А НЕ ТРЕМЯ, ПОТОМУ ЧТО ДИПСИК ВСЁ НАПУТАЛ
// НО Я ТЕПЕРЬ ХОТЯ БЫ РАЗОБРАЛСЯ
// НУЖНО ОТПРАВЛЯТЬ ОБРАТНО (с помощью d.Nack(false, true)?) В ГЛАВНУЮ ОЧЕРЕДЬ,
// ИНКРЕМЕНТИРУЯ СЧЕТЧИК И ЭКСПОНЕНЦИАЛЬНО УВЕЛИЧИВАЯ ВРЕМЯ ОБРАБОТКИ - в гроке посмотреть
// А ПОТОМ ПРОСТО В DLX ОТПРАВЛЯТЬ

type AMTConn struct {
	Transport    *MQTransport
	ExchangeName string
	QueueName    string
}

func NewAMTConn(t *MQTransport, Name string, PrefetchCount int) (*AMTConn, error) {
	ExchangeName := "Exchange_" + Name
	QueueName := "Queue_" + Name
	// Основной обменник
	err := t.ch.ExchangeDeclare(ExchangeName, "direct", true, false, false, false, nil)
	if err != nil {
		return nil, errors.New("error in AMTConn func: Main exchange declare: " + err.Error())
	}

	// DLX для ошибок
	err = t.ch.ExchangeDeclare(ExchangeName+"_DLX", "direct", true, false, false, false, nil)
	if err != nil {
		return nil, errors.New("error in AMTConn func: DLX exchange declare: " + err.Error())
	}

	// DLQ (очередь для ошибочных сообщений)
	_, err = t.ch.QueueDeclare(QueueName+"_DLQ", true, false, false, false, nil)
	if err != nil {
		return nil, errors.New("error in AMTConn func: DLQ queue declare: " + err.Error())
	}
	err = t.ch.QueueBind(QueueName+"_DLQ", QueueName+"_DLQ_key", ExchangeName+"_DLX", false, nil)
	if err != nil {
		return nil, errors.New("error in AMTConn func: DLQ queue bind: " + err.Error())
	}

	// Очередь для повторных отправок (этакий отстойник) с таймаутом 15 секунд до возврата в основную очередь
	args := amqp091.Table{
		"x-dead-letter-exchange":    ExchangeName,             // Обменник для ошибок
		"x-dead-letter-routing-key": QueueName + "_retry_key", // Ключ для повторной отправки
		"x-message-ttl":             15000,                    // 15 сек задержка
	}
	_, err = t.ch.QueueDeclare(QueueName+"_retry", true, false, false, false, args)
	if err != nil {
		return nil, errors.New("error in AMTConn func: Retry queue declare: " + err.Error())
	}
	err = t.ch.QueueBind(QueueName+"_retry", QueueName+"_retry_key", ExchangeName, false, nil)
	if err != nil {
		return nil, errors.New("error in AMTConn func: Retry queue bind: " + err.Error())
	}

	// Основная очередь - тут проверяется заголовок с количеством повторных отправлений
	// и чуть что - отправляется в DLQ
	args = amqp091.Table{
		"x-dead-letter-exchange":    ExchangeName + "_DLX",  // Обменник для ошибок
		"x-dead-letter-routing-key": QueueName + "_DLQ_key", // Ключ для DLQ
	}
	_, err = t.ch.QueueDeclare(QueueName, true, false, false, false, args)
	if err != nil {
		return nil, errors.New("error in AMTConn func: Main queue declare: " + err.Error())
	}
	err = t.ch.QueueBind(QueueName, QueueName+"_key", ExchangeName, false, nil)
	if err != nil {
		return nil, errors.New("error in AMTConn func: Main queue bind: " + err.Error())
	}

	err = t.ch.Qos(PrefetchCount, 0, false)
	if err != nil {
		return nil, errors.New("error in AMTConn func: Channel QoS setting: " + err.Error())
	}

	return &AMTConn{Transport: t, ExchangeName: ExchangeName, QueueName: QueueName}, nil
}

func (c *AMTConn) Publish(msg []byte) error {
	return c.Transport.ch.Publish(
		c.ExchangeName,
		c.QueueName+"_key",
		false,
		false,
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent,
			ContentType:  "application/json",
			Body:         msg,
			Headers:      amqp091.Table{"retry-count": 0}, // Счетчик попыток
		},
	)
}

func (c *AMTConn) Consume() (<-chan amqp091.Delivery, error) {
	msg, err := c.Transport.ch.Consume(c.QueueName, "", false, false, false, false, nil)
	if err != nil {
		// залогировать
		return nil, errors.New("error during consuming message: " + err.Error())
	}
	return msg, nil
}

func (c *AMTConn) Close() error {
	if err := c.Transport.Close(); err != nil {
		return err
	}
	return nil
}

func (c *AMTConn) Retry(msg amqp091.Delivery) error {
	retries, ok := msg.Headers["retry-count"].(int)
	if !ok {
		msg.Headers["retry-count"] = -1
		msg.Nack(false, false) // Отправка в DLQ
		return nil
	}
	if retries > 2 { // Превышен лимит попыток или ошибка
		msg.Nack(false, false) // Отправка в DLQ
		return nil
	}
	msg.Headers["retry-count"] = retries + 1
	err := c.Transport.ch.Publish(
		c.ExchangeName,
		c.QueueName+"_retry_key",
		false,
		false,
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent,
			ContentType:  msg.ContentType,
			Body:         msg.Body,
			Headers:      msg.Headers,
		},
	)
	if err != nil {
		msg.Headers["retry-count"] = -2
		msg.Nack(false, false)
		return errors.New("error during publishing to retry queue: " + err.Error())
	}
	return msg.Ack(false)
}
