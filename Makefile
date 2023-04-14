build:
	@rm -rf release/
	@go mod tidy
	@echo "Building for Linux (AMD64) & MacOS (ARM64)..."
	@GOOS=linux GOARCH=amd64 go build -o release/kraken-lin64/ ./...
	@GOOS=darwin GOARCH=arm64 go build -o release/kraken-macARM64/ ./...
	@echo "Done building"

clean:
	@rm -rf release/
	@rm -rf debug/
	@rm -rf logs/
	@go mod tidy
	@docker stop kraken
	@docker rm kraken
	@clear

postgres:
	@docker run --name kraken -e POSTGRES_USER=kraken -e POSTGRES_PASSWORD=kraken -p 5432:5432 -v /var/lib/postgresql/data -d postgres

dev:
	@rm -rf debug/
	@rm -rf logs/
	@go mod tidy
	@clear
	@go build -o debug/ ./...
	@./debug/kraken -c "./files/config.dev.json"

diag:
	@clear
	@go build -o debug/ ./...
	@./debug/kraken -c "./files/config.dev.json" -d