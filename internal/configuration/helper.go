package configuration

import (
	"errors"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

func LoadEnv(cfg *AMTConfig) {
	(*cfg).Transport.DSN = os.Getenv("AMT_TRANSPORT_DSN")
	(*cfg).Queue.Name = os.Getenv("AMT_QUEUE_NAME")
	countstr := os.Getenv("AMT_QOS_PREFETCHCOUNT")
	count, err := strconv.Atoi(countstr)
	if err != nil {
		(*cfg).QoS.PrefetchCount = 0
		return
	}
	(*cfg).QoS.PrefetchCount = count
}

func LoadEtcd(cfg *AMTConfig, etcdURL, file string) error {
	err := viper.AddRemoteProvider("etcd3", etcdURL, file)
	if err != nil {
		return errors.New("AddRemoteProvider error: " + err.Error())
	}
	extension, err := findExtension(file)
	if err != nil {
		return err
	}
	viper.SetConfigType(extension)
	if err := viper.ReadRemoteConfig(); err != nil {
		return errors.New("Error reading config file: " + err.Error())
	}
	// прочитать конфигурацию в структуру
	(*cfg).Transport.DSN = viper.GetString("amt.transport.dsn")
	(*cfg).Queue.Name = viper.GetString("amt.queue.name")
	(*cfg).QoS.PrefetchCount = viper.GetInt("amt.qos.prefetchcount")
	return nil
}

func findExtension(file string) (string, error) {
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '.' {
			return string(file[i+1:]), nil
		}
	}
	return "", errors.New("no extension found in config file: ")
}
