package valkey

import (
	"log/slog"
	"os"

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
	logger *slog.Logger,
	url string,
) valkey.Client {
	client, err := NewValkeyClient(url)
	if err != nil {
		logger.Error("failed to connect to valkey",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
	return client
}
