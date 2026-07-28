.PHONY: hooks lint test check

hooks:
	pre-commit install --install-hooks --hook-type pre-commit --hook-type pre-push --hook-type commit-msg

lint:
	pre-commit run --all-files --hook-stage pre-commit

test:
	go test ./...

check:
	pre-commit run --all-files --hook-stage pre-commit
	pre-commit run --all-files --hook-stage pre-push
