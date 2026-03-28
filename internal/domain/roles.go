package domain

type (
	Role string
)

const (
	ADMIN     Role = "ADMIN"
	ATTENDANT Role = "ATTENDANT"
	MECHANIC  Role = "MECHANIC"
	CLIENT    Role = "CLIENT"
)
