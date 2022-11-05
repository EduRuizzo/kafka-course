package config

import (
	"context"
	"log"

	"github.com/sethvargo/go-envconfig"
)

type Client struct {
	SeedBrokers []string `env:"SEED_BROKERS,default=[::1]:9092"`
}

func MustNewClient() Client {
	c := Client{}
	if err := envconfig.Process(context.Background(), &c); err != nil {
		log.Fatal(err)
	}

	return c
}
