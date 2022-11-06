package pipeconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/EduRuizzo/kafka-course/model"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type Consumer struct {
	kcl    *kgo.Client
	logger *zap.Logger
	close  chan bool
	group  string
}

func NewConsumer(kcl *kgo.Client, group string) *Consumer {
	l, _ := zap.NewProduction()

	co := &Consumer{
		kcl:    kcl,
		logger: l.Named(group),
		close:  make(chan bool),
		group:  group,
	}

	return co
}

func (co *Consumer) Close() {
	co.logger.Info("closing consumer")
	co.close <- true
	co.kcl.Close()
	co.logger.Info("consumer closed correctly")
}

func (co *Consumer) PollFetches(ctx context.Context) {
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
			co.logger.Info("consume loop gracefully closed", zap.String("context", ctxt.Err().Error()))

			return
		case fet := <-fc:
			fet.EachPartition(func(p kgo.FetchTopicPartition) {
				for _, record := range p.Records {
					var val model.BasicPayload

					err := json.Unmarshal(record.Value, &val)
					if err != nil {
						log.Println("error unmarshalling record value")
						continue
					}
					fmt.Printf("Key: %s, Value %+v, from range inside a callback!\n", record.Key, val)
				}
			})

			err := co.kcl.CommitUncommittedOffsets(ctx)
			if err != nil {
				co.logger.Error("coudln't commit offsets", zap.Error(err))
			}

			cancel()
		}
	}
}
