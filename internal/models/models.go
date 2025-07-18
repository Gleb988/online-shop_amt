package models

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
