.PHONY: build lint test check e2e-tests watch format dev-e2e-test-env dev-e2e-test-env-rebuild

lua_config=./configs/nvim/lua.editorconfig

build:
	CGO_ENABLED=0 go build -o bin/mycli ./mycli

lint:
	golangci-lint run ./...

test:
	go test ./... -timeout 60m

check:
	make build
	make lint
	make test

e2e-tests:
	go test -tags="integration" ./... -timeout 60m

watch:
	modd

format:
	golines --max-len=120 --base-formatter="gofumpt" -w .
	find ./configs/nvim -iname '*.lua' | \
		xargs -I {} CodeFormat format -c $(lua_config) -f {} -ow

dev-e2e-test-env:
	go run ./test/cmd
	docker start -i system_setup_dev

dev-e2e-test-env-rebuild:
	-docker stop system_setup_dev
	-docker rm system_setup_dev
	-docker rmi system_setup_dev_img
