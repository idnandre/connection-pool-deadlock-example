FROM golang:1.22-bullseye as builder
WORKDIR /app
RUN apt-get update && apt-get -y install \
    libssl-dev \
    build-essential
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -o connection-pool-deadlock-example main.go

FROM debian:bullseye-slim
COPY --from=builder /app/connection-pool-deadlock-example /
CMD ["/connection-pool-deadlock-example"]