SHELL := /bin/bash

run:
	go run app/sales-api/main.go

runadmin:
	go run app/sales-admin/main.go
	
tidy:
	go mod tidy
	go mod vendor

test:
	go test -v ./... -count=1
	#staticcheck ./...