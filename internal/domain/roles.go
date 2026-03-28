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

var (
	AttendantRoles            = []Role{ATTENDANT, ADMIN}
	MechanicRoles             = []Role{MECHANIC, ADMIN}
	ClientRoles               = []Role{CLIENT, ADMIN}
	AttendantAndMechanicRoles = []Role{ATTENDANT, MECHANIC, ADMIN}
)
