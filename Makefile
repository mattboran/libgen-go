build:
	go get -d ./... && go build -o bin/libgen libgen.go

run:
	go run libgen.go
