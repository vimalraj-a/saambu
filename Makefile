.PHONY: help install install-server install-web \
	server web dev \
	build build-server build-web \
	clean

help:
	@echo "Targets:"
	@echo "  install         Install dependencies for server and web"
	@echo "  server          Run the Go API server"
	@echo "  web             Run the React dev server"
	@echo "  dev             Run server and web together"
	@echo "  build           Build production artifacts for server and web"
	@echo "  clean           Remove build artifacts"

install: install-server install-web

install-server:
	cd server && go mod download

install-web:
	cd web && npm install

server:
	cd server && go run .

web:
	cd web && npm run dev

dev:
	@trap 'kill 0' EXIT INT TERM; \
	$(MAKE) server & \
	$(MAKE) web & \
	wait

build: build-server build-web

build-server:
	cd server && go build -o bin/server .

build-web:
	cd web && npm run build

clean:
	rm -rf server/bin web/dist
