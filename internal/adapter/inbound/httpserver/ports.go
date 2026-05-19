package httpserver

import "context"

// HealthChecker is satisfied by any type that can report database liveness.
type HealthChecker interface {
	Ping(ctx context.Context) error
}
