package middleware

import "net/http"

const (
	AdminRole     = "ADMIN"
	AttendantRole = "ATTENDANT"
	MechanicRole  = "MECHANIC"
	ClientRole    = "CLIENT"
)

var publicRoutes = map[string]struct{}{
	"/v1/auth/login":    {},
	"/v1/auth/register": {},
	"/ping":             {},
}

var privateRoutes = map[string]map[string][]string{
	http.MethodPost: {
		"/v1/customers": {AdminRole, AttendantRole},
		"/v1/services":  {AdminRole, AttendantRole},
	},
	http.MethodGet: {
		"/v1/services": {AdminRole, AttendantRole},
	},
	http.MethodPut: {
		"/v1/services/:id": {AdminRole, AttendantRole},
	},
	http.MethodDelete: {
		"/v1/services/:id": {AdminRole},
	},
}
