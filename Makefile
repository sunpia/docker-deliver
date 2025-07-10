build:
	go build -o ./dist/docker-deliver ./cmd/docker-deliver

install:
	go install ./cmd/docker-deliver