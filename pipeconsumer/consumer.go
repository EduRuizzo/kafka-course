package pipeconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/EduRuizzo/kafka-course/model"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

const pollLoopPeriod = 20 * time.Millisecond
const pollTimeout = 5 * time.Second

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
	co.logger.Info("consumer correctly initialized")

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

	tick := time.NewTicker(pollLoopPeriod)

	for {
		select {
		case <-co.close:
			co.logger.Info("consume loop gracefully closed")
			return
		case <-tick.C:
			ctxt, c := context.WithTimeout(ctx, pollTimeout)
			fet := co.kcl.PollFetches(ctxt)
			select {
			case <-ctx.Done(): // continue loop if no fetches within timeout
			default:
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
			}
			c()
		}
	}
}
