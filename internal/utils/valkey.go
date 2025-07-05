package utils

import (
	"github.com/rs/zerolog/log"
	"github.com/valkey-io/valkey-go"
)

func NewValkeyClient(url string) valkey.Client {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{url},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to valkey")
	}
	return client
}
