package jobs

import (
	"context"
	"fmt"

	"github.com/alexburley/ask-howard/internal/port/outbound"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type Client struct {
	river *river.Client[pgx.Tx]
}

var _ outbound.JobEnqueuer = (*Client)(nil)

func NewClient(ctx context.Context, pool *pgxpool.Pool, workers *river.Workers) (*Client, error) {
	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Workers: workers,
		Queues:  map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: 4}},
	})
	if err != nil {
		return nil, fmt.Errorf("create river client: %w", err)
	}
	return &Client{river: riverClient}, nil
}

func (c *Client) Start(ctx context.Context) error {
	if err := c.river.Start(ctx); err != nil {
		return fmt.Errorf("start river: %w", err)
	}
	return nil
}

func (c *Client) Stop(ctx context.Context) error {
	if err := c.river.Stop(ctx); err != nil {
		return fmt.Errorf("stop river: %w", err)
	}
	return nil
}

func (c *Client) EnqueueExtraction(ctx context.Context, setID, userID uuid.UUID) error {
	_, err := c.river.Insert(ctx, ExtractionArgs{SetID: setID, UserID: userID}, nil)
	if err != nil {
		return fmt.Errorf("insert extraction job: %w", err)
	}
	return nil
}
