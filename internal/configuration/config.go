package configuration

import (
	"errors"
	"os"
)

type AMTConfig struct {
	Transport struct {
		DSN string
	}
	Queue struct {
		Name string
	}
	QoS struct {
		PrefetchCount int
	}
}

// single structure with all parameters
func LoadConfig(file string) (AMTConfig, error) {
	cfg := AMTConfig{}
	if etcd := os.Getenv("ETCD_URL"); etcd != "" {
		if file == "" {
			return AMTConfig{}, errors.New("CONFIG_FILE must be given in Flow parameters")
		}
		if err := LoadEtcd(&cfg, etcd, file); err != nil {
			return AMTConfig{}, errors.New("error during LoadEtcd: " + err.Error())
		}
	} else {
		LoadEnv(&cfg)
	}

	if cfg.Transport.DSN == "" {
		return AMTConfig{}, errors.New("AMT_TRANSPORT_DSN must be set")
	}
	if cfg.Queue.Name == "" {
		return AMTConfig{}, errors.New("AMT_QUEUE_NAME must be set")
	}
	if cfg.QoS.PrefetchCount < 1 {
		return AMTConfig{}, errors.New("AMT_QOS_PREFETCHCOUNT must be > 1")
	}

	return cfg, nil
}
