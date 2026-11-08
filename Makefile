SHELL := /bin/sh

.PHONY: help validate-template setup format lint test build security ci backend-test backend-lint backend-build

help:
	@printf '%s\n' 'Targets:' \
		'  make validate-template - Run repository structure and link validation' \
		'  make backend-test      - Run Go backend tests' \
		'  make backend-lint      - Run Go vet' \
		'  make backend-build     - Compile Go backend' \
		'  make test              - Run all Python and Go tests' \
		'  make lint              - Run Python and Go static checks' \
		'  make build             - Compile Python scripts and Go backend' \
		'  make security          - Run repository validation checks' \
		'  make ci                - Run the full repository and backend validation suite'

validate-template:
	python3 scripts/validate_repo.py
	python3 -m unittest discover -s tests -v

format:
	python3 -m py_compile scripts/*.py tests/*.py
	cd backend && gofmt -w $$(find . -name '*.go' -type f)

lint: backend-lint
	python3 -m py_compile scripts/*.py tests/*.py
	python3 scripts/validate_repo.py

backend-test:
	cd backend && go test -race ./...

backend-lint:
	cd backend && go vet ./...

backend-build:
	cd backend && go build ./...

test: backend-test
	python3 -m unittest discover -s tests -v

build: backend-build
	python3 -m py_compile scripts/*.py

security:
	python3 scripts/validate_repo.py

ci: validate-template lint test build security
