all: crabcore ssh kubernetes

VERSION_PKG := github.com/zerodotfive/crabcore/internal/version
VERSION := $(shell git describe --tags --exact-match 2>/dev/null || git rev-parse HEAD)

uninstall:
	rm -rf ~/.local/share/crabcore ~/.local/bin/crabcore ~/.crabcore.yaml ~/.config/systemd/user/crabcore* ~/.config/systemd/user/*/crabcore* ~/.local/share/systemd/timers; systemctl --user daemon-reload

go-mod-download:
	go mod download

crabcore: go-mod-download
	go build -ldflags "-X $(VERSION_PKG).CrabcoreVersion=$(VERSION)" -o ./bin/crabcore ./cmd/crabcore

example: go-mod-download
	go build --buildmode=plugin -o ./lib/example.so ./plugins/example

ssh: go-mod-download
	go build --buildmode=plugin -o ./lib/ssh.so ./plugins/ssh

kubernetes: go-mod-download
	go build --buildmode=plugin -o ./lib/kubernetes.so ./plugins/kubernetes
