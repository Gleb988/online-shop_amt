package configuration

import (
	"errors"
)

type AMTConfig struct {
	MQ struct {
		Transport struct {
			DSN string
		}
		Queue struct {
			Name       string
			DLQName    string
			Durable    bool
			AutoDelete bool
			Exclusive  bool
			NoWait     bool
		}
		QoS struct {
			PrefetchCount int
			PrefetchSize  int
			Global        bool
		}
		Consume struct {
		}
		Rublish struct {
		}
	}
	AMT struct {
		MessageProcessTimeout int
		MessageWorkers        int
	}
}

// single structure with all parameters
func LoadConfig() (AMTConfig, error) {
	cfg := AMTConfig{}

	cfg.MQ.Transport.DSN = GetMQDSN("amqp://guest:guest@mq:5672/")

	cfg.MQ.Queue.Name = GetMQQueueName("message")
	cfg.MQ.Queue.DLQName = GetMQDLQName("messageDLQ")
	cfg.MQ.Queue.Durable = GetMQQueueDurable(true)
	cfg.MQ.Queue.AutoDelete = GetMQQueueAutodelete(false)
	cfg.MQ.Queue.Exclusive = GetMQQueueExclusive(false)
	cfg.MQ.Queue.NoWait = GetMQQueueNowait(false)

	cfg.MQ.QoS.PrefetchCount = GetMQQoSPrefetchCount(1)
	cfg.MQ.QoS.PrefetchSize = GetMQQoSPrefetchSize(0)
	cfg.MQ.QoS.Global = GetMQQoSGlobal(false)

	cfg.AMT.MessageProcessTimeout = GetMessageProcessTimeout(5)
	cfg.AMT.MessageWorkers = GetMessageWorkers(1)

	if cfg.MQ.QoS.PrefetchCount < 1 {
		return AMTConfig{}, errors.New("MQ_QOS_PREFETCH_COUNT must be > 0")
	}
	if cfg.MQ.QoS.PrefetchSize < 0 {
		return AMTConfig{}, errors.New("MQ_QOS_PREFETCH_SIZE must be >= 0")
	}
	if cfg.AMT.MessageProcessTimeout < 1 {
		return AMTConfig{}, errors.New("AMT_MESSAGE_PROCESS_TIMEOUT must be > 0")
	}
	if cfg.AMT.MessageWorkers < 1 {
		return AMTConfig{}, errors.New("AMT_MESSAGE_WORKERS_QUANTITY must be > 0")
	}

	return cfg, nil
}
