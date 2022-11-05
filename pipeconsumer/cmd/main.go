package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/EduRuizzo/kafka-course/config"
	"github.com/EduRuizzo/kafka-course/pipeconsumer"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	var group, topic string

	// gracefully exit on keyboard interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	flag.StringVar(&group, "g", "kgo-group", "name of the consumer group")
	flag.StringVar(&topic, "t", "getting-started", "kafka topic")
	flag.Parse()

	cfg := config.MustNewClient()

	kcl, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.SeedBrokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		kcl.Close()
		log.Panic(err)
	}

	pco := pipeconsumer.NewConsumer(kcl, group)
	pco.PollFetches(context.Background())

	<-c
	pco.Close()
	os.Exit(0)
}
