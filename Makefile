.PHONY : build
build:
	CC=x86_64-linux-musl-gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "-linkmode external -extldflags -static -s -w" -trimpath -o ./VaultBacker main.go