FROM golang:1.13

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD app -port ${PORT:-:443} -domain ${DOMAIN:-www.example.com}