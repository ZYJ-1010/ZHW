.PHONY: dev-go dev-funds dev-admin test build smoke

dev-go:
	cd services/go-api && go run ./cmd/server

dev-funds:
	cd services/funds-service && javac -encoding UTF-8 -d target/classes src/main/java/com/zhw/funds/FundsApplication.java src/main/java/com/zhw/funds/common/*.java && java -cp target/classes com.zhw.funds.FundsApplication

dev-admin:
	cd admin-web && npm run dev

test:
	cd services/go-api && go test ./...
	cd admin-web && npm run build

build:
	cd services/go-api && go build ./cmd/server
	cd admin-web && npm run build

smoke:
	curl http://127.0.0.1:8080/health
	curl http://127.0.0.1:8081/health
