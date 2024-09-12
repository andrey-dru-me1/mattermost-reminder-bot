# build stage
FROM golang:1.23.1-alpine AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -v -o /go/bin/app .

# final stage
FROM alpine:3.20.3

RUN apk add --no-cache curl

HEALTHCHECK --interval=10m --timeout=10s --start-period=5s --retries=5 CMD [ "curl", "--fail", "localhost:8080/healthcheck" ]

RUN addgroup -S app && adduser -S app -G app
COPY --from=builder --chown=app /go/bin/app /app
USER app

EXPOSE 80

ENTRYPOINT ["/app"]
