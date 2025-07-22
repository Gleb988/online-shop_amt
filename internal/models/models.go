package models

type AMTConfig struct {
	Transport struct {
		DSN string
	}
	ExchangeConfig struct {
		Name string
	}
	QueueConfig struct {
		Name    string
		Durable bool
	}
	ChannelQoS struct {
		PrefetchCount int
	}
}
