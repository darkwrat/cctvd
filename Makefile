all:
	protoc -I cctv --go_out=plugins=grpc:./cctv ./cctv/cctv.proto
	go build -o ./bin/cctvc ./cmd/cctvc/main.go
	GOOS=linux GOARCH=arm go build -o ./bin/cctvc-arm ./cmd/cctvc/main.go
	go build -o ./bin/cctvd ./cmd/cctvd/main.go
	GOOS=linux GOARCH=arm go build -o ./bin/cctvd-arm ./cmd/cctvd/main.go
