lua_format=~/.cache/nvim/myconfig/lua_lsp/3rd/EmmyLuaCodeStyle/build/CodeFormat/CodeFormat
lua_config=./configs/nvim/lua.editorconfig

build:
	CGO_ENABLED=0 go build -o bin/mycli ./mycli

unit-tests:
	go test ./... -timeout 60m

e2e-tests:
	go test -tags="integration" ./... -timeout 60m

watch:
	modd

format:
	golines --max-len=120 --base-formatter="gofumpt" -w .
	find ./configs/nvim -iname '*.lua' | \
		xargs -I {} $(lua_format) format -c $(lua_config) -f {} -ow

dev-e2e-test-env:
	go run ./test/cmd
	docker start -i system_setup_dev

dev-e2e-test-env-rebuild:
	-docker stop system_setup_dev
	-docker rm system_setup_dev
	-docker rmi system_setup_dev_img
