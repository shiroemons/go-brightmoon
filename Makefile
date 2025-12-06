.PHONY: build test lint lint-fix fmt vet clean

# ビルド
build:
	go build -v ./...

build-brightmoon:
	go build -o brightmoon ./cmd/brightmoon

build-titles:
	go build -o titles_th ./cmd/titles_th

# テスト
test:
	go test -v -race ./...

test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

test-cover-html:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# リント
lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v2.7.1 golangci-lint run ./...

lint-fix:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v2.7.1 golangci-lint run --fix ./...

# フォーマット
fmt:
	go fmt ./...

# 静的解析
vet:
	go vet ./...

# 依存関係
mod-verify:
	go mod verify

mod-tidy:
	go mod tidy

# クリーンアップ
clean:
	rm -f brightmoon titles_th coverage.out coverage.html
	rm -rf dist/ tmp/
