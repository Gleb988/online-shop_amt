package configuration

import (
	"os"
	"strconv"
	"strings"

	"github.com/Gleb988/online-shop_amt/internal/models"
)

func (c AMTConfig) ToQueueConfig() models.QueueConfig {
	return models.QueueConfig{
		QName:      c.MQ.Queue.Name,
		DLQName:    c.MQ.Queue.DLQName,
		Durable:    c.MQ.Queue.Durable,
		AutoDelete: c.MQ.Queue.AutoDelete,
		Exclusive:  c.MQ.Queue.Exclusive,
		NoWait:     c.MQ.Queue.NoWait,
	}
}

func (c AMTConfig) ToQoSCongif() models.ChannelQoS {
	return models.ChannelQoS{
		PrefetchCount: c.MQ.QoS.PrefetchCount,
		PrefetchSize:  c.MQ.QoS.PrefetchSize,
		Global:        c.MQ.QoS.Global,
	}
}

// key: MQ_DSN
func GetMQDSN(defaultDSN string) string {
	dsn := os.Getenv("MQ_DSN")
	if dsn == "" {
		return defaultDSN
	}
	return dsn
}

// key: MQ_QUEUE_NAME
func GetMQQueueName(defaultQueueName string) string {
	name := os.Getenv("MQ_QUEUE_NAME")
	if name == "" {
		return defaultQueueName
	}
	return name
}

// key: MQ_QUEUE_DLQ_NAME
func GetMQDLQName(defaultQueueName string) string {
	name := os.Getenv("MQ_QUEUE_DLQ_NAME")
	if name == "" {
		return defaultQueueName
	}
	return name
}

// key: MQ_QUEUE_DURABLE
func GetMQQueueDurable(dflt bool) bool {
	state := os.Getenv("MQ_QUEUE_DURABLE")
	return strings.TrimSpace(strings.ToLower(state)) == "true"
}

// key: MQ_QUEUE_AUTODELETE
func GetMQQueueAutodelete(dflt bool) bool {
	state := os.Getenv("MQ_QUEUE_AUTODELETE")
	return strings.TrimSpace(strings.ToLower(state)) == "true"
}

// key: MQ_QUEUE_EXCLUSIVE
func GetMQQueueExclusive(dflt bool) bool {
	state := os.Getenv("MQ_QUEUE_EXCLUSIVE")
	return strings.TrimSpace(strings.ToLower(state)) == "true"
}

// key: MQ_QUEUE_NOWAIT
func GetMQQueueNowait(dflt bool) bool {
	state := os.Getenv("MQ_QUEUE_NOWAIT")
	return strings.TrimSpace(strings.ToLower(state)) == "true"
}

// key: MQ_QOS_PREFETCH_COUNT
func GetMQQoSPrefetchCount(defaultQueueLength int) int {
	lengthStr := os.Getenv("MQ_QOS_PREFETCH_COUNT")
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return -1
	}
	return length
}

// key: MQ_QOS_PREFETCH_SIZE
func GetMQQoSPrefetchSize(dflt int) int {
	strsize := os.Getenv("MQ_QOS_PREFETCH_SIZE")
	size, err := strconv.Atoi(strsize)
	if err != nil {
		return -1
	}
	return size
}

// key: MQ_QOS_GLOBAL
func GetMQQoSGlobal(dflt bool) bool {
	state := os.Getenv("MQ_QOS_GLOBAL")
	return strings.TrimSpace(strings.ToLower(state)) == "true"
}

// key: AMT_MESSAGE_PROCESS_TIMEOUT
func GetMessageProcessTimeout(defaultTimeout int) int {
	timeoutStr := os.Getenv("AMT_MESSAGE_PROCESS_TIMEOUT")
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return 0
	}
	return timeout
}

// key: AMT_MESSAGE_WORKERS_QUANTITY
func GetMessageWorkers(defaultWorkers int) int {
	quantityStr := os.Getenv("AMT_MESSAGE_WORKERS_QUANTITY")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		return 0
	}
	return quantity
}
