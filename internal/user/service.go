package user

import "log/slog"

type Service struct {
	logger *slog.Logger
	users  Users
}

func NewService(l *slog.Logger, us Users) *Service {
	return &Service{
		logger: l.With("svc", "user"),
		users:  us,
	}
}
