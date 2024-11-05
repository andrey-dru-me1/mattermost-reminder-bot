# build stage
FROM golang:1.23.2-alpine3.20 AS builder

RUN apk add --no-cache git
WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -v -o /go/bin/app .

# final stage
FROM alpine:3.20.3

RUN addgroup -S app && adduser -S app -G app
COPY --from=builder --chown=app /go/bin/app /app
USER app

ENTRYPOINT ["/app"]
