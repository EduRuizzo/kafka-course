package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"flag"

	"github.com/EduRuizzo/kafka-course/config"
	"github.com/EduRuizzo/kafka-course/model"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	to = 3
	nr = 10
)

func main() {
	topic, key, name := "", "", ""

	var nRec int

	var tls, sasl bool

	flag.StringVar(&topic, "t", "getting-started", "kafka topic")
	flag.StringVar(&key, "k", "myKey", "kafka key")
	flag.StringVar(&name, "n", "TRANSACTION", "name for the record value")
	flag.BoolVar(&tls, "tls", false, "TLS enabled or disabled")
	flag.BoolVar(&sasl, "sasl", false, "SASL auth enabled or disabled")
	flag.IntVar(&nRec, "nrec", nr, "num of records to include in transaction")
	flag.Parse()
	os.Setenv("SERVER_CERT_FILE", "../../certs/server.cer.pem")

	cfg := config.MustNewClientConfig()

	opts := config.ClientSecurityConfigToKafkaClientOpts(&cfg, tls, sasl)
	opts = append(opts,
		kgo.TransactionalID("PRODUCER-TRANSACTIONAL-ID"),
		kgo.DefaultProduceTopic(topic),
		kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelInfo, func() string {
			return "[input producer] "
		})),
	)

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		log.Panic(err)
	}
	defer cl.Close()

	ctx, cancel := context.WithTimeout(context.Background(), to*time.Second)
	defer cancel()

	err = cl.Ping(ctx)
	if err != nil {
		log.Panic(err)
	}

	// TRANSACTION
	if errt := cl.BeginTransaction(); errt != nil {
		log.Panic("couldn't begin transaction: ", errt)
	}

	for i := 0; i < nRec; i++ {
		val := model.BasicPayload{
			Name: fmt.Sprintf("%s-%v", name, i),
			UUID: uuid.NewString(),
			Date: time.Now(),
		}

		v, errm := json.Marshal(val)
		if errm != nil {
			log.Println("error marshaling payload:", errm)
			continue
		}

		record := &kgo.Record{Topic: topic, Key: []byte(key), Value: v}

		if errp := cl.ProduceSync(ctx, record).FirstErr(); errp != nil {
			log.Printf("record had a produce error while synchronously producing: %v\n", errp)
		}
	}

	ctxb := context.Background()

	err = cl.Flush(ctxb)
	if err != nil {
		log.Println("error flushing transaction:", err)
	}

	commit := kgo.TransactionEndTry(err == nil)

	switch err := cl.EndTransaction(ctxb, commit); err {
	case nil:
		log.Printf("transaction finished correctly, %v records produced\n", nRec)
	case kerr.OperationNotAttempted:
		if erra := cl.EndTransaction(ctxb, kgo.TryAbort); erra != nil {
			log.Panic("abort failed: ", erra)
		}
	default:
		log.Panic("commit failed: ", err)
	}
}
