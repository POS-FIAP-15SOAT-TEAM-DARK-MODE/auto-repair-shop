package main

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/factory"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/routing"
)

func main() {
	container := factory.HttpContainer()
	router := routing.SetupRouter(container)
	router.Run(":8080")
}
