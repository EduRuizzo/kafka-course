package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	seeds := []string{"localhost:9092"}
	// One client can both produce and consume!
	// Consuming can either be direct (no consumer group), or through a group. Below, we use a group.
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		// kgo.ConsumerGroup("my-group-identifier"),
		// kgo.ConsumeTopics("foo"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer cl.Close()
	ctx := context.Background()

	// 1.) Producing a message
	// All record production goes through Produce, and the callback can be used
	// to allow for synchronous or asynchronous production.
	var wg sync.WaitGroup
	wg.Add(1)
	record := &kgo.Record{Topic: "getting-started", Key: []byte("myKey"), Value: []byte(time.Now().String())}
	cl.Produce(ctx, record, func(_ *kgo.Record, err error) {
		defer wg.Done()
		if err != nil {
			log.Printf("record had a produce error: %v\n", err)
		}

	})
	wg.Wait()

	// Alternatively, ProduceSync exists to synchronously produce a batch of records.
	if err := cl.ProduceSync(ctx, record).FirstErr(); err != nil {
		log.Printf("record had a produce error while synchronously producing: %v\n", err)
	}
}
