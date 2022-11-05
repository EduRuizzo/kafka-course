package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"flag"

	"github.com/EduRuizzo/kafka-course/config"
	"github.com/EduRuizzo/kafka-course/model"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

const to = 3

func main() {
	topic, key, name := "", "", ""
	flag.StringVar(&topic, "t", "getting-started", "kafka topic")
	flag.StringVar(&key, "k", "myKey", "kafka key")
	flag.StringVar(&name, "n", "JoJo", "name for the record value")
	flag.Parse()

	cfg := config.MustNewClient()

	// One client can both produce and consume!
	// Consuming can either be direct (no consumer group), or through a group. Below, we use a group.
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.SeedBrokers...),
		// kgo.ConsumerGroup("my-group-identifier"),
		// kgo.ConsumeTopics("foo"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer cl.Close()

	ctx, cancel := context.WithTimeout(context.Background(), to*time.Second)
	defer cancel()

	err = cl.Ping(ctx)
	if err != nil {
		log.Panic(err)
	}

	// 1.) Producing a message
	// All record production goes through Produce, and the callback can be used
	// to allow for synchronous or asynchronous production.
	var wg sync.WaitGroup

	wg.Add(1)

	val := model.BasicPayload{
		Name: name,
		UUID: uuid.NewString(),
		Date: time.Now(),
	}

	v, err := json.Marshal(val)
	if err != nil {
		log.Panic("error marshaling payload:", err)
	}

	record := &kgo.Record{Topic: topic, Key: []byte(key), Value: v}
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
