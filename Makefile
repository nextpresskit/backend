# NextPressKit — three public Make targets. Everything else: ./scripts/nextpresskit help

.DEFAULT_GOAL := default

.PHONY: default setup run postman-sync \
	install deps tidy clean build build-all \
	migrate-up migrate-drop db-fresh seed \
	test test-coverage test-integration security-check checks \
	start stop

default:
	@echo "NextPressKit — use:"
	@echo "  make setup         text menu (TTY) or NP_SETUP_NONINTERACTIVE=1 for linear bootstrap only"
	@echo "  make run           API in the foreground"
	@echo "  make postman-sync  Refresh gitignored postman/ from templates + .env"
	@echo "  make migrate-up    Run DB migrations"
	@echo "  make seed          Seed demo data"
	@echo "  make db-fresh      Drop+recreate public schema (dev only)"
	@echo "Advanced: ./scripts/nextpresskit help"

setup:
	@bash scripts/nextpresskit setup

run:
	@bash scripts/nextpresskit run

postman-sync:
	@bash scripts/nextpresskit postman-sync

install:
	@bash scripts/nextpresskit install

deps:
	@bash scripts/nextpresskit deps

tidy:
	@bash scripts/nextpresskit tidy

clean:
	@bash scripts/nextpresskit clean

build:
	@bash scripts/nextpresskit build

build-all:
	@bash scripts/nextpresskit build-all

migrate-up:
	@bash scripts/nextpresskit migrate-up

migrate-drop:
	@bash scripts/nextpresskit migrate-drop

db-fresh:
	@bash scripts/nextpresskit db-fresh

seed:
	@bash scripts/nextpresskit seed

test:
	@bash scripts/nextpresskit test

test-coverage:
	@bash scripts/nextpresskit test-coverage

test-integration:
	@bash scripts/nextpresskit test-integration

security-check:
	@bash scripts/nextpresskit security-check

checks:
	@bash scripts/nextpresskit checks

start:
	@bash scripts/nextpresskit start

stop:
	@bash scripts/nextpresskit stop
