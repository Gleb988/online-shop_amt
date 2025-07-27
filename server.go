package amt

import (
	"context"
	"errors"
	"sync"
	"time"

	// КОНФИГ НУЖЕН, ПОТОМУ ЧТО ИНАЧЕ ПРИДЕТСЯ В ДВУХ МЕСТАХ ПРОПИСЫВАТЬ ВСЕ ПЕРЕМЕННЫЕ
	// А ТАК ПРИ ЗАПУСКЕ СЕРВЕРА КОНФИГУРАЦИИ ВСЁ ЗАПУСТИЛ - И ОК
	// ЛИБО ДАЖЕ ПОСЛЕ МОЖНО ДОБАВИТЬ КОНФИГ ФАЙЛЫ ДРУГИМИ НОВЫМИ СКРИПТАМИ
	"github.com/Gleb988/online-shop_amt/internal/configuration"
	"github.com/Gleb988/online-shop_amt/internal/transport"
	"github.com/rabbitmq/amqp091-go"
)

type AppAPI interface {
	ProcessMessage(context.Context, amqp091.Delivery) error
}

type AMT interface {
	Publish(msg []byte) error
	Consume() (<-chan amqp091.Delivery, error)
	Retry(msg amqp091.Delivery) error
	Close() error
}

// можно вынести wg, workerPool и timeout, поскольку они нужны только в RunConsumer
type Flow struct {
	MQ  AMT
	App AppAPI
	//wg         *sync.WaitGroup
	//workerPool chan struct{}
	//timeout    time.Duration
}

//var ErrNack = ErrNackType{Message: "AMT ErrNack! Do not send to retry queue"}

// В КОНФИГЕ НЕ БУДЕТ ЖЕСТКОЙ СВЯЗИ С etcd.
// ЕСЛИ НЕ УКАЗАН URL ДЛЯ НЕГО, ТО ИСПОЛЬЗУЮТСЯ ENV
// Переменная file = "", когда переменные из окружения (только для одного потока пока)
// Только для одного потока пока
func NewFlowWithAutoConfig(app AppAPI, file string) (*Flow, error) {
	cfg, err := configuration.LoadConfig(file)
	if err != nil {
		return nil, errors.New("error during configuration: " + err.Error())
	}
	conn, err := transport.NewMQTransport(cfg.Transport.DSN)
	if err != nil {
		return nil, errors.New("error during transport creating: " + err.Error())
	}
	mq, err := transport.NewAMTConn(conn, cfg.Queue.Name, cfg.QoS.PrefetchCount)
	if err != nil {
		return nil, errors.New("error during queue creating: " + err.Error())
	}

	return &Flow{MQ: mq, App: app}, nil
}

// AddFlowWithAutoConfig создает новый Flow для работы с очередью сообщений, используя параметры, указанные в файле конфигурации, хранящемся в etcd.
//
// Параметры:
//   - app: интерфейс API приложения
//   - conn: MQ-транспорт
//   - file: имя файла (должно быть уникальным)
//
// Возвращает:
//   - *Flow: объект Flow
//   - error: ошибка, если не удалось создать соединение
//
// Примечания:
//   - Перед использованием необходимо вызвать amt.Dial().
//   - Имена очередей должны быть уникальными.
func AddFlowWithAutoConfig(app AppAPI, conn *transport.MQTransport, file string) (*Flow, error) {
	cfg, err := configuration.LoadConfig(file)
	if err != nil {
		return nil, errors.New("error during configuration: " + err.Error())
	}
	mq, err := transport.NewAMTConn(conn, cfg.Queue.Name, cfg.QoS.PrefetchCount)
	if err != nil {
		return nil, errors.New("error during queue creating: " + err.Error())
	}
	return &Flow{MQ: mq, App: app}, nil
}

// NewFlow создает новый Flow для работы с очередью сообщений.
//
// Параметры:
//   - app: интерфейс API приложения
//   - conn: MQ-транспорт
//   - name: имя очереди (должно быть уникальным)
//   - prefetchCount: количество предзагружаемых сообщений (должно быть > 0)
//
// Возвращает:
//   - *Flow: объект Flow
//   - error: ошибка, если не удалось создать соединение
//
// Примечания:
//   - Перед использованием необходимо вызвать amt.Dial().
//   - Имена очередей должны быть уникальными.
func NewFlow(app AppAPI, conn *transport.MQTransport, name string, prefetchCount int) (*Flow, error) {
	if prefetchCount < 1 {
		return nil, errors.New("prefetchCount must be > 0")
	}
	mq, err := transport.NewAMTConn(conn, name, prefetchCount)
	if err != nil {
		return nil, errors.New("error during queue creating: " + err.Error())
	}

	return &Flow{MQ: mq, App: app}, nil
}

func Dial(dsn string) (*transport.MQTransport, error) {
	t, err := transport.NewMQTransport(dsn)
	if err != nil {
		return nil, errors.New("error during transport creating: " + err.Error())
	}
	return t, nil
}

// тут НЕ нужно использовать конфиг, потому что это относится уже к пользовательской стороне, а не к настройке "путей сообщения" или сервера
func (s *Flow) RunConsumer(ctx context.Context, workers int, timeout time.Duration) (err error) {
	defer func() {
		var errs []error
		errs = append(errs, err)
		if err := s.MQ.Close(); err != nil {
			errs = append(errs, err)
		}
		err = errors.Join(errs...)
	}()

	var wg sync.WaitGroup
	workerPool := make(chan struct{}, workers)
	errChan := make(chan error, workers)

	msgs, err := s.MQ.Consume()
	if err != nil {
		return errors.New("error during consuming message: " + err.Error())
	}
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				wg.Wait()
				return errors.New("msgs channel closed: " + err.Error())
			}
			workerPool <- struct{}{}
			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
					<-workerPool
				}()
				ctx, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()
				err := s.App.ProcessMessage(ctx, msg)
				if err != nil {
					var nackErr ErrNack
					if errors.As(err, &nackErr) {
						msg.Nack(false, false)
						return
					}
					err := s.MQ.Retry(msg)
					if err != nil {
						errChan <- err
					}
				}
				msg.Ack(false)
			}()
		case <-ctx.Done():
			wg.Wait()
			return errors.New("msgs channel closed: " + ctx.Err().Error())
		case err := <-errChan:
			wg.Wait()
			return err
		}
	}
}
