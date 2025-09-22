package valkey

import (
	"context"
	"fmt"

	valkeygo "github.com/valkey-io/valkey-go"
)

type Counter struct {
	client valkeygo.Client
}

func NewCounter(
	client valkeygo.Client,
) *Counter {
	return &Counter{
		client: client,
	}
}

func (c *Counter) Incr(ctx context.Context) (int, error) {
	const op = "repository.valkey.Counter.Incr"

	key := c.buildKey()
	cmd := c.client.B().Incr().Key(key).Build()
	result := c.client.Do(ctx, cmd)
	if result.Error() != nil {
		return 0, fmt.Errorf("%s: %w", op, result.Error())
	}

	count, err := result.AsInt64()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return int(count), nil
}

func (c *Counter) buildKey() string {
	return "counters:short_code"
}
