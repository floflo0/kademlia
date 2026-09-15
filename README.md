# Kademlia

## Setup project

### Generate protobuf code

First install [protobuf compiler](https://protobuf.dev/installation/) and the [go
plugin](https://protobuf.dev/getting-started/gotutorial/).

Then, generate the required protobuf files with:

```sh
go generate ./...
```

## Run

To run the application:

```sh
go run -v ./cmd/kademlia
```

## Build

To build the application:

```sh
go build -v -o kademlia ./cmd/kademlia
```

## Docker

Launch multitple instances using docker:

```sh
docker compose up --build
```

## Static Analysis

Run the static analysis to detect errors in go files:

```sh
go vet -v ./...
```

## Formatting

To format all files:

```sh
go fmt ./...
```

## Testing

To run tests:

```sh
go test -v -cover -race ./...
```
