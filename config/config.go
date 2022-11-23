package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"

	"github.com/sethvargo/go-envconfig"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

type ClientConfig struct {
	ClientCertFile  string   `env:"CLIENT_CERT_FILE,default=../../certs/client.cer.pem"`
	ClientKeyFile   string   `env:"CLIENT_KEY_FILE,default=../../certs/client.key.pem"`
	ServerCertFile  string   `env:"SERVER_CERT_FILE,default=../../certs/server.cer.pem"`
	SASLUser        string   `env:"SASL_USER,default=edu"`
	SASLPassword    string   `env:"SASL_PASSWORD,default=edu-secret"`
	SeedBrokers     []string `env:"SEED_BROKERS,default=[::1]:9092"`
	SSLSeedBrokers  []string `env:"SSL_SEED_BROKERS,default=[::1]:9093"`
	SASLSeedBrokers []string `env:"SASL_SEED_BROKERS,default=[::1]:9094"`
}

func MustNewClientConfig() ClientConfig {
	c := ClientConfig{}
	if err := envconfig.Process(context.Background(), &c); err != nil {
		log.Fatal(err)
	}

	return c
}

func NewTLSConfig(clientCertFile, clientKeyFile, caCertFile string, clientAuth bool) (*tls.Config, error) {
	/* #nosec G402 */
	tlsConfig := tls.Config{
		MinVersion:         tls.VersionTLS12,
		NextProtos:         []string{"SSL"},
		InsecureSkipVerify: true,
		ServerName:         "[::1]", // Can´t be an IP, this doesn´t serve, so we enable Insecure to test.
	}

	if clientAuth {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
		if err != nil {
			return &tlsConfig, err
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA cert
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		return &tlsConfig, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	return &tlsConfig, err
}

func ClientSecurityConfigToKafkaClientOpts(cfg *ClientConfig, tlsOpt, sasl bool) []kgo.Opt {
	opts := []kgo.Opt{}

	if tlsOpt { // Mutual TLS Authentication
		tlsCfg, err := NewTLSConfig(cfg.ClientCertFile, cfg.ClientKeyFile, cfg.ServerCertFile, true)
		if err != nil {
			log.Panic(err)
		}

		opts = append(opts, kgo.DialTLSConfig(tlsCfg), kgo.SeedBrokers(cfg.SSLSeedBrokers...))

		return opts
	}

	if sasl { // SASL SCRAM
		tlsCfg, err := NewTLSConfig(cfg.ClientCertFile, cfg.ClientKeyFile, cfg.ServerCertFile, false)
		if err != nil {
			log.Panic(err)
		}

		opts = append(opts,
			kgo.SASL(
				scram.Auth{
					User: cfg.SASLUser,
					Pass: cfg.SASLPassword,
				}.AsSha512Mechanism()),
			kgo.DialTLSConfig(tlsCfg),
			kgo.SeedBrokers(cfg.SASLSeedBrokers...),
		)

		return opts
	}

	opts = append(opts, kgo.SeedBrokers(cfg.SeedBrokers...))

	return opts
}
