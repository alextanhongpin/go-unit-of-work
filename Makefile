include .env
export


start:
	@go run main.go


up:
	@docker-compose up -d


down:
	@docker-compose down


test:
	@go test -v -race -cpuprofile=cpu.out -memprofile=mem.out -coverprofile=cov.out
