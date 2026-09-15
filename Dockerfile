FROM golang:1.27-alpine3.24 AS builder

WORKDIR /app

RUN apk add --no-cache protoc

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

COPY go.mod go.sum ./

RUN go mod download

COPY proto proto

RUN go generate -v -x ./...

COPY . .

RUN CGO_ENABLED=0 go build -v -o kademlia ./cmd/kademlia

FROM scratch

WORKDIR /app

COPY --from=builder /app/kademlia /app/kademlia

CMD ["./kademlia"]
