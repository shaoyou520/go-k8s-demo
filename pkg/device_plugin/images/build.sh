#/bin/bash

CGO_ENABLED=0  GOOS=linux  GOARCH=amd64  go build -o my-device ../main.go
docker build -t qtdocker/my-device -f Dockerfile .
docker push qtdocker/my-device

