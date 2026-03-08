package domain

type PingService interface {
	Execute() Ping
}

type Ping struct {
	Message string
}
