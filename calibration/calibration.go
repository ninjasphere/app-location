package calibration

import "github.com/ninjasphere/go-ninja/api"

type Service struct {
	conn *ninja.Connection
}

type Status struct {
	Progress int
}

func NewService(conn *ninja.Connection) *Service {
	return &Service{}
}

type Location struct {
	ID      string
	Name    string
	Quality int
}
