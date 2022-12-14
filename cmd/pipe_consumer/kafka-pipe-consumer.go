package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/EduRuizzo/kafka-course/config"
	"github.com/EduRuizzo/kafka-course/pipeconsumer"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

func main() {
	var group, topic string

	var tls, sasl, transactional bool

	// gracefully exit on keyboard interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	flag.StringVar(&group, "g", "kgo-group", "name of the consumer group")
	flag.StringVar(&topic, "t", "getting-started", "kafka topic")
	flag.BoolVar(&tls, "tls", false, "TLS enabled or disabled")
	flag.BoolVar(&sasl, "sasl", false, "SASL auth enabled or disabled")
	flag.BoolVar(&transactional, "tr", false, "transactional consumer")
	flag.Parse()

	cfg := config.MustNewClientConfig()

	opts := config.ClientSecurityConfigToKafkaClientOpts(&cfg, tls, sasl)
	opts = append(opts,
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		kgo.DisableAutoCommit(),
	)

	if transactional { // only read messages that have been written as part of committed transactions
		opts = append(opts,
			kgo.FetchIsolationLevel(kgo.ReadCommitted()),
			kgo.RequireStableFetchOffsets(),
		)
	}

	kcl, err := kgo.NewClient(opts...)
	if err != nil {
		kcl.Close()
		log.Panic(err)
	}

	pco := pipeconsumer.NewConsumer(kcl, group)
	l, _ := zap.NewProduction()
	l.Info("kafka consumer created", zap.String("group", group), zap.String("topic", topic), zap.String("host", strings.Join(cfg.SeedBrokers, ",")))

	pco.Start(context.Background())

	<-c // Wait for interrupt to gracefully close
	pco.Close()
}
