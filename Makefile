gen_private_key:
	openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:4096

gen_public_key:
	openssl rsa -pubout -in private.pem -out public.pem

run:
	@echo "Start app"
	go run ./cmd/gophermart/main.go

build:
	@echo "Build app"
	go build -o ./cmd/gophermart/main.go
all:
	@echo "Start build app"

.PHONY: cover
cover:
	go test -short -count=1 -coverprofile=coverage.out ./internal/... ./cmd/...
	go tool cover -html=coverage.out
	rm ./coverage.out
migrate:
	@echo "Start migrate up"
	migrate -path migrations -database "postgres://user:password@localhost:5432/keeper?sslmode=disable" up

down:
	@echo "Start migrate down"
	migrate -path migrations -database "postgres://user:password@localhost:5432/keeper?sslmode=disable" down

#// migrate create -ext sql -dir migrations -seq create_users_table
#  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
# migrate -path migrations -database "postgres://user:password@localhost:5432/loyalty?sslmode=disable" up

# mockery --name=Repository --output=mocks --outpkg=mocks