FROM golang:1.16-alpine

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go build -o app .

CMD ["/go/src/app/app"]
