FROM golang:alpine as go
WORKDIR /app
ENV GO111MODULE=on

COPY go.mod .
RUN go mod download

COPY . .
RUN go build -o server-discovery ./cmd

FROM alpine

WORKDIR /app

COPY --from=go /app/server-discovery ./server-discovery
COPY run/config.yaml ./config.yaml
CMD ["./server-discovery"]