# Helpers
IS_DARWIN := $(filter Darwin,$(shell uname -s))

define set_env
	sed $(if $(IS_DARWIN),-i "",-i) -e "s/^#*\($(1)=\).*/$(if $(2),,#)\1$(2)/" .env
endef

EXEC := docker compose exec -T hidra
EXEC_TTY := docker compose exec hidra

# Environment recipes
.PHONY: default
default: init up

.PHONY: init
init:
	test -f .env || cp .env.example .env
	$(call set_env,USER_ID,$(shell id -u))
	mkdir -p ~/go/pkg

.PHONY: up
up:
	DOCKER_BUILDKIT=1 docker compose up -d --build

.PHONY: down
down:
	docker compose down

.PHONY: shell
shell:
	$(EXEC_TTY) zsh

# Project recipes
.PHONY: deps
deps:
	$(EXEC) go mod download

.PHONY: run
run:
	$(EXEC_TTY) go run main.go exporter /etc/hidra_exporter/exporter.yml

.PHONY: debug
debug:
	$(EXEC_TTY) dlv debug --listen=:2345 --headless --api-version=2 main.go

.PHONY: lint
lint:
	$(EXEC) go vet ./...

.PHONY: test
test:
	$(EXEC) go test -v ./...

.PHONY: sync-git-prehooks
sync-git-prehooks:
	cp -r githooks/* .git/hooks/
	chmod +x .git/hooks/*

.PHONY: clean-release
clean-release:
	rm -rf build
