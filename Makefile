.PHONY: install lint test contract evals check

install:
	pip install -e ".[dev]"

lint:
	ruff check src tests
	mypy src

test:
	pytest tests/unit -q

contract:
	pytest tests/contract -q

evals:
	pytest tests/evals -q

check: lint test contract evals
