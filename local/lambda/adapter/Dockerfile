# HTTP adapter in front of the login lambda (local ecosystem only). Standard library only.
FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/adapter main.go

FROM scratch
COPY --from=builder /out/adapter /adapter
EXPOSE 9001
ENTRYPOINT ["/adapter"]
