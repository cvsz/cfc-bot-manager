SHELL := /bin/sh

.PHONY: help validate-template setup format lint test build security ci

help:
	@printf '%s\n' 'Targets:' \
		'  make validate-template - Run repository structure and link validation' \
		'  make test              - Run python unit tests' \
		'  make lint              - Run python syntax and structure checks' \
		'  make build             - Run compilation check on scripts' \
		'  make security          - Run security and validation checks' \
		'  make ci                - Run full CI validation suite'

validate-template:
	python3 scripts/validate_repo.py
	python3 -m unittest discover -s tests -v

format:
	python3 -m py_compile scripts/*.py tests/*.py

lint:
	python3 -m py_compile scripts/*.py tests/*.py
	python3 scripts/validate_repo.py

test:
	python3 -m unittest discover -s tests -v

build:
	python3 -m py_compile scripts/*.py

security:
	python3 scripts/validate_repo.py

ci: validate-template lint test build security
