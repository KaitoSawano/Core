all:
		go build -o xcosh xcosh.go
		go build -o xcosh-cli ./cmd/xcosh-cli.go

install: all
		sudo mv xcosh /usr/local/bin/
		sudo mv xcosh-cli /usr/local/bin/
