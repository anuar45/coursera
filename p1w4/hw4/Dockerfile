FROM golang:1.12-alpine

WORKDIR /go/src/github.com/anuar45/coursera/

COPY . .

CMD CGO_ENABLED=0 go test -v