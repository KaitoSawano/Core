all:
	go build -o xcosh xcosh.go
	go build -o xcosh-cli ./cmd/xcosh-cli.go

install: all
	mv xcosh /usr/local/bin/
	mv xcosh-cli /usr/local/bin/
