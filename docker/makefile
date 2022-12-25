all: app container

app: 
	go build -o ../bin/echoserver ../main.go

container:
	cp ../bin/echoserver ./
	docker build --tag dachunky/echoserver:latest -f echoserver.Dockerfile .