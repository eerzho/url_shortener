package valkey

import (
	"log"

	"github.com/valkey-io/valkey-go"
)

func NewValkeyClient(url string) (valkey.Client, error) {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{url},
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func MustNewValkeyClient(url string) valkey.Client {
	client, err := NewValkeyClient(url)
	if err != nil {
		log.Fatalf("failed to connect to valkey: %v\n", err)
	}
	return client
}
