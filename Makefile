# Update the version to your needs via env.
BUILD_VERSION = $(KRAKEN_BUILD_VERSION)
BUILD_DATE=$$(date +%Y.%m.%d-%H:%M:%S)

# DO NOT CHANGE.
build:
	@[ "${KRAKEN_BUILD_VERSION}" ] || ( echo "KRAKEN_BUILD_VERSION is not set"; exit 1 )
	@echo "Building app for Windows (AMD64), Linux (AMD64) & MacOS (ARM64)..."
	@go mod tidy
	@GOOS=windows GOARCH=amd64 go build -v -trimpath -ldflags="-s -w -X github.com/devusSs/twitch-kraken/internal/updater.buildVersion=${BUILD_VERSION} -X github.com/devusSs/twitch-kraken/internal/updater.buildDate=${BUILD_DATE}" -o release/kraken_win_amd64/ ./...
	@GOOS=linux GOARCH=amd64 go build -v -trimpath -ldflags="-s -w -X github.com/devusSs/twitch-kraken/internal/updater.buildVersion=${BUILD_VERSION} -X github.com/devusSs/twitch-kraken/internal/updater.buildDate=${BUILD_DATE}" -o release/kraken_lin_amd64/ ./...
	@GOOS=darwin GOARCH=arm64 go build -v -trimpath -ldflags="-s -w -X github.com/devusSs/twitch-kraken/internal/updater.buildVersion=${BUILD_VERSION} -X github.com/devusSs/twitch-kraken/internal/updater.buildDate=${BUILD_DATE}" -o release/kraken_mac_arm64/ ./...
	@echo "Done building app"

# DO NOT CHANGE.
clean:
	@clear
	@go mod tidy
	@rm -rf ./debug/
	@rm -rf ./release/
	@rm -rf ./dist/
	@rm -rf ./logs/
	@rm -rf ./tmp/
	@rm -rf ./testing/

# DO NOT CHANGE.
dev: build
	@clear
	@rm -rf ./testing
	@mkdir ./testing
	@mkdir ./testing/files
	@cp -R ./files ./testing
	@cp ./release/kraken_mac_arm64/kraken ./testing
	@cd ./testing && ./kraken -c "./files/config.dev.json" -su

# DO NOT CHANGE.
diag: build
	@clear
	@-mkdir ./testing
	@-mkdir ./testing/files
	@cp -R ./files ./testing
	@cp ./release/kraken_mac_arm64/kraken ./testing
	@cd ./testing && ./kraken -c "./files/config.dev.json" -d

# DO NOT CHANGE.
secure-cookie: build
	@clear
	@-mkdir ./testing
	@-mkdir ./testing/files
	@cp -R ./files ./testing
	@cp ./release/kraken_mac_arm64/kraken ./testing
	@cd ./testing && ./kraken -sc