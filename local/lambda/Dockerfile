# Local image of the customer login lambda (auto-repair-shop-lambda-auth).
# Built with the lambda repository as the build context (services/auto-repair-shop-lambda-auth, see docker-compose.yml).
# The public provided:al2023 image ships the Lambda Runtime Interface Emulator (RIE): when it runs outside AWS it
# serves POST /2015-03-31/functions/function/invocations on port 8080, which the adapter calls.
FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/bootstrap ./cmd/lambda

FROM public.ecr.aws/lambda/provided:al2023
# Custom runtime: the handler is the executable named "bootstrap" in the runtime directory.
COPY --from=builder /out/bootstrap /var/runtime/bootstrap
CMD ["bootstrap"]
