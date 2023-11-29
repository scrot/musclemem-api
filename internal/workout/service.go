package workout

import "log/slog"

// Service represents a workout service
type Service struct {
	logger   *slog.Logger
	workouts Workouts
}

// NewService returns a new workout service that
// can interact with the workouts datastore
func NewService(l *slog.Logger, ws Workouts) *Service {
	return &Service{
		logger:   l.With("svc", "workout"),
		workouts: ws,
	}
}
