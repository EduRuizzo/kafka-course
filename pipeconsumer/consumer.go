package pipeconsumer

import (
	"context"
	"encoding/json"

	"github.com/EduRuizzo/kafka-course/model"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

const pipeSize = 10 // max amount of fetches before polling thread block

type Consumer struct {
	kcl    *kgo.Client
	logger *zap.Logger
	close  chan bool
	fetc   chan kgo.Fetches
	group  string
}

func NewConsumer(kcl *kgo.Client, group string) *Consumer {
	l, _ := zap.NewProduction()

	co := &Consumer{
		kcl:    kcl,
		logger: l.Named(group),
		close:  make(chan bool),
		group:  group,
		fetc:   make(chan kgo.Fetches, pipeSize),
	}

	return co
}

func (co *Consumer) Start(ctx context.Context) {
	go co.processFetches(co.fetc) // processor thread
	go co.pollFetches(ctx)        // polling thread
}

func (co *Consumer) Close() {
	co.logger.Info("closing consumer")
	co.close <- true
	co.kcl.Close()
	co.logger.Info("consumer closed correctly")
}

// Poll fetches is the polling thread
func (co *Consumer) pollFetches(ctx context.Context) {
	co.logger.Info("consumer poll loop started")

	for {
		ctxt, cancel := context.WithCancel(ctx)
		fc := make(chan kgo.Fetches)

		go func() {
			fet := co.kcl.PollFetches(ctxt)
			fc <- fet
		}()
		select {
		case <-co.close:
			cancel()
			co.logger.Info("poll loop gracefully closed", zap.String("context", ctxt.Err().Error()))
			close(co.fetc) // close channel so process loop ends also

			return
		case fet := <-fc:
			co.fetc <- fet // send fetch to the processor thread

			cancel()
		}
	}
}

func (co *Consumer) processFetches(fetc <-chan kgo.Fetches) {
	co.logger.Info("consumer process loop started")

	for fet := range fetc {
		fet.EachPartition(func(p kgo.FetchTopicPartition) {
			for _, rec := range p.Records {
				var val model.BasicPayload

				err := json.Unmarshal(rec.Value, &val)
				if err != nil {
					co.logger.Error("error unmarshalling record value", zap.Error(err))
					continue
				}
				co.logger.Info("Record processed", zap.String("key", string(rec.Key)), zap.Any("value", val))
			}
		})

		err := co.kcl.CommitUncommittedOffsets(context.Background())
		if err != nil {
			co.logger.Error("coudln't commit offsets", zap.Error(err))
		}
	}

	co.logger.Info("process loop gracefully closed")
}
