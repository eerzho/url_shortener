package utils

import (
	"log"

	"github.com/valkey-io/valkey-go"
)

func NewValkeyClient(url string) valkey.Client {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{url},
	})
	if err != nil {
		log.Fatalf("failed to connect to valkey: %v", err)
	}

	return client
}
