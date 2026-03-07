package main

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/routing"

func main() {
	router := routing.SetupRouter()
	router.Run(":8080")
}
