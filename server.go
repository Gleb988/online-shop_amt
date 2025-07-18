package amt

import (
	"context"
	"errors"
	"sync"
	"time"

	configuration "github.com/Gleb988/online-shop_amt/internal/configuratioin"
	"github.com/Gleb988/online-shop_amt/internal/transport"
	"github.com/rabbitmq/amqp091-go"
)

type AppAPI interface {
	ProcessMessage(context.Context, []byte) error
}

type ConsumeCloser interface {
	Consume() (<-chan amqp091.Delivery, error)
	Close() error
}

type Server struct {
	MQ         ConsumeCloser
	App        AppAPI
	wg         *sync.WaitGroup
	workerPool chan struct{}
	timeout    time.Duration
}

func NewServer(app AppAPI, workers int, timeout time.Duration) (*Server, error) {
	var wg sync.WaitGroup
	pool := make(chan struct{}, workers)

	cfg, err := configuration.LoadConfig()
	if err != nil {
		return nil, errors.New("error during configuration: " + err.Error())
	}
	conn, err := transport.NewMQTransport(cfg.MQ.Transport.DSN)
	if err != nil {
		return nil, errors.New("error during transport creating: " + err.Error())
	}
	mq, err := transport.NewMQQueue(conn, cfg.ToQueueConfig(), cfg.ToQoSCongif())
	if err != nil {
		return nil, errors.New("error during queue creating: " + err.Error())
	}

	return &Server{MQ: mq, App: app, wg: &wg, workerPool: pool, timeout: timeout}, nil
}

func (s *Server) RunConsumer(ctx context.Context) error {
	defer func() {
		if err := s.MQ.Close(); err != nil {
			// залогировать?
		}
	}()
	msgs, err := s.MQ.Consume()
	if err != nil {
		return errors.New("error during consuming message: " + err.Error())
	}
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				s.wg.Wait()
				return errors.New("msgs channel closed: " + err.Error())
			}
			s.workerPool <- struct{}{}
			s.wg.Add(1)
			go func() {
				defer func() {
					s.wg.Done()
					<-s.workerPool
				}()
				ctx, cancel := context.WithTimeout(ctx, s.timeout)
				defer cancel()
				err := s.App.ProcessMessage(ctx, msg.Body)
				if err != nil {
					msg.Nack(false, false) // НАСТРОИТЬ
				}
			}()
		case <-ctx.Done():
			s.wg.Wait()
			return errors.New("msgs channel closed: " + err.Error())
		}
	}
}
