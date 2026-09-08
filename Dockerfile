FROM golang:1.27-alpine3.24 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -v -o kademlia ./cmd/kademlia

FROM scratch

WORKDIR /app

COPY --from=builder /app/kademlia /app/kademlia

CMD ["./kademlia"]
