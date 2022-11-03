package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/EduRuizzo/kafka-course/model"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	seeds := []string{"[::1]:9092"}
	// One client can both produce and consume!
	// Consuming can either be direct (no consumer group), or through a group. Below, we use a group.
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup("kgo-group"),
		kgo.ConsumeTopics("getting-started"),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer cl.Close()

	ctx := context.Background()

	// 2.) Consuming messages from a topic
	for {
		fetches := cl.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			// All errors are retried internally when fetching, but non-retriable errors are
			// returned from polls so that users can notice and take action.
			panic(fmt.Sprint(errs))
		}

		// We can iterate through a record iterator...
		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			fmt.Println(string(record.Value), "from an iterator!")
		}

		// or a callback function.
		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			for _, record := range p.Records {
				var val model.BasicPayload

				err := json.Unmarshal(record.Value, &val)
				if err != nil {
					log.Println("error unmarshalling record value")
					continue
				}
				fmt.Printf("Key: %s, Value %+v, from range inside a callback!\n", record.Key, val)
			}

			// We can even use a second callback!
			p.EachRecord(func(record *kgo.Record) {
				var val model.BasicPayload
				err := json.Unmarshal(record.Value, &val)
				if err != nil {
					log.Println("error unmarshalling record value")
				} else {
					fmt.Printf("Key: %s, Value %+v, from a second callback!\n", record.Key, val)
				}
			})
		})
		err = cl.CommitUncommittedOffsets(ctx)
		if err != nil {
			log.Fatal("couldn´t commit offsets", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
