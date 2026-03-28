# Auth middleware — how to use

This document explains how to attach the Auth middleware to routes and which role constants you can use.

Summary
- The project exposes a reusable middleware named `Auth` (package `internal/infra/middleware`) that validates the incoming JWT and enforces RBAC.
- Roles are defined in `internal/domain/roles.go`. There are single-role constants and commonly used role slices.

Quick examples

- Per-route (require specific roles)

```go
import (
  role "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
  "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/middleware"
)

// require ATTENDANT or ADMIN
v1.POST("/customers", middleware.Auth(role.AttendantRoles...), c.CustomerHandler.Create())

// require ATTENDANT or MECHANIC or ADMIN
v1.GET("/services", middleware.Auth(role.AttendantAndMechanicRoles...), c.WorkHandler.List())
```

- Per-group (apply the same requirement to many routes)

```go
protected := v1.Group("/", middleware.Auth(role.AttendantRoles...))
protected.POST("/services", c.WorkHandler.Create())
protected.POST("/vehicle", c.VehicleHandler.Create())
```

- Require authentication only (no specific role)

If you want to require a valid JWT but no particular role, call the middleware without role arguments (or pass an empty slice):

```go
v1.POST("/some-protected", middleware.Auth(), handler)
```

Available roles and helpers
- Single-role constants (type `domain.Role`):
  - `domain.ADMIN`
  - `domain.ATTENDANT`
  - `domain.MECHANIC`
  - `domain.CLIENT`

- Convenience role slices (already defined):
  - `domain.AttendantRoles`            -> `{ATTENDANT, ADMIN}`
  - `domain.MechanicRoles`             -> `{MECHANIC, ADMIN}`
  - `domain.ClientRoles`               -> `{CLIENT, ADMIN}`
  - `domain.AttendantAndMechanicRoles` -> `{ATTENDANT, MECHANIC, ADMIN}`

Notes and best practices
- Use route groups when many endpoints share the same access rules — this keeps wiring simple and avoids duplicate calls.
- The middleware reads the JWT claims and checks roles from the token. Make sure tokens include a `roles` claim (the app's `GenerateToken` stores roles in the token).