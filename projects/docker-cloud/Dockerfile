# use the latest official golang image
FROM golang:latest

# set the working directory in the container
WORKDIR /usr/src/docker-cloud

# copy the go.mod and go.sum to the working directory
COPY go.mod go.sum ./

# download and verify the dependencies
RUN go mod download && go mod verify

# copy the main.go and main_test.go to the working directory
COPY main.go main_test.go ./

# build the go application and put the binary in /usr/local/bin
RUN go build -v -o /usr/local/bin/docker-cloud ./...

# set a port environment variable with a default value
ENV HTTP_PORT=80

# expose the port specified above
EXPOSE ${HTTP_PORT}

# run the binary when the container starts
CMD ["docker-cloud"]