FROM golang:latest

WORKDIR /usr/src/docker-cloud

COPY go.mod ./

RUN go mod download && go mod verify

COPY . .

RUN go build -v -o /usr/local/bin/docker-cloud ./...

ENV HTTP_PORT=80

EXPOSE ${HTTP_PORT}

CMD ["docker-cloud"]
