debug:
	DEBUG=1 go run main.go

release:
	go build -o ./build/l8zykube main.go

