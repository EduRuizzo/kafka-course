package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"

	"github.com/sethvargo/go-envconfig"
	"github.com/twmb/franz-go/pkg/kgo"
)

type ClientConfig struct {
	ClientCertFile string   `env:"CLIENT_CERT_FILE,default=../../certs/client.cer.pem"`
	ClientKeyFile  string   `env:"CLIENT_KEY_FILE,default=../../certs/client.key.pem"`
	ServerCertFile string   `env:"SERVER_CERT_FILE,default=../../certs/server.cer.pem"`
	SeedBrokers    []string `env:"SEED_BROKERS,default=[::1]:9092"`
	SSLSeedBrokers []string `env:"SEED_BROKERS,default=[::1]:9093"`
}

func MustNewClientConfig() ClientConfig {
	c := ClientConfig{}
	if err := envconfig.Process(context.Background(), &c); err != nil {
		log.Fatal(err)
	}

	return c
}

func NewTLSConfig(clientCertFile, clientKeyFile, caCertFile string) (*tls.Config, error) {
	/* #nosec G402 */
	tlsConfig := tls.Config{
		MinVersion:         tls.VersionTLS12,
		NextProtos:         []string{"SSL"},
		InsecureSkipVerify: true,
		ServerName:         "[::1]",
	}

	// Load client cert
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		return &tlsConfig, err
	}

	tlsConfig.Certificates = []tls.Certificate{cert}

	// Load CA cert
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		return &tlsConfig, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	// tlsConfig.BuildNameToCertificate()

	return &tlsConfig, err
}

func ClientConfigToKafkaClientOpts(cfg *ClientConfig, group, topic string, tlsOpt bool) []kgo.Opt {
	opts := []kgo.Opt{
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		kgo.DisableAutoCommit(),
	}

	if tlsOpt {
		tlsCfg, err := NewTLSConfig(cfg.ClientCertFile, cfg.ClientKeyFile, cfg.ServerCertFile)
		if err != nil {
			log.Panic(err)
		}

		opts = append(opts, kgo.DialTLSConfig(tlsCfg), kgo.SeedBrokers(cfg.SSLSeedBrokers...))

		return opts
	}

	opts = append(opts, kgo.SeedBrokers(cfg.SeedBrokers...))

	return opts
}
