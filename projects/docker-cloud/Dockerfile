FROM golang:latest

WORKDIR /usr/src/docker-cloud

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -v -o /usr/local/bin/docker-cloud ./...

ENV HTTP_PORT=80

EXPOSE ${HTTP_PORT}

CMD ["docker-cloud"]

# https://hub.docker.com/repository/docker/bazmurphy/docker-cloud