ARG GO_VERSION=1.18.3

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN apk --no-cache add build-base ca-certificates
RUN CGO_ENABLED=1 GOOS=linux go build -o /app -a -ldflags '-linkmode external -extldflags "-static"' ./cmd/api

FROM scratch AS dev

COPY --from=builder /app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/app", "--cert=/var/lib/certs/localhost.cert.pem", "--key=/var/lib/certs/localhost.key.pem"]
