package ping

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"

type service struct{}

func Service() *service {
	return &service{}
}

func (s *service) Execute() domain.Ping {
	return domain.Ping{Message: "pong"}
}
