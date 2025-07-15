package valkey

import (
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

func MustNewValkeyClient(
	url string,
) valkey.Client {
	client, err := NewValkeyClient(url)
	if err != nil {
		panic(err)
	}
	return client
}
