# Security

The gateway is the only public port of the system. Its job is to forward credentials untouched and to leak nothing about the internals.

## Rules

- The gateway does not validate tokens and does not read a database. Each service validates its own JWT. `Authorization` is forwarded as is.
- Never log or return the `Authorization` header, a token, a password, a CPF or a CNPJ. The logger redacts sensitive field names, but code must not pass them in the first place.
- The query string is not logged: it can carry a customer document (`GET /v1/customers?document=...`).
- `/internal/*` is blocked at the edge and never forwarded, whatever the method. Service-to-service REST is not part of the public API.
- Upstream errors are mapped to fixed messages (`upstream timeout`, `upstream unavailable`). The cause (addresses, refused connections, DNS names) stays in the server log.
- Panics answer a generic `internal server error`; the panic value and stack go to the server log only.
- Client-supplied `X-Request-Id` is accepted only when it has at most 128 characters from `A-Z a-z 0-9 - _ . :`, so logs cannot be forged with line breaks.
- Upstream URLs must be absolute `http` or `https` URLs. Invalid values are rejected at startup without echoing them.
- The metrics `method` label is bounded: unknown methods are reported as `OTHER`.
