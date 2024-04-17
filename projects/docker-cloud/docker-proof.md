# docker proof

`docker build . -t docker-cloud`

```sh
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/docker-cloud$ docker build . -t docker-cloud
[+] Building 5.1s (11/11) FINISHED docker:default
=> [internal] load build definition from Dockerfile 0.0s
=> => transferring dockerfile: 549B 0.0s
=> [internal] load metadata for docker.io/library/golang:latest 0.3s
=> [internal] load .dockerignore 0.0s
=> => transferring context: 2B 0.0s
=> [1/6] FROM docker.io/library/golang:latest@sha256:450e3822c7a135e1463cd83e51c8e2eb03b86a02113c89424e6f0f8344bb4168 0.0s
=> [internal] load build context 0.0s
=> => transferring context: 1.22kB 0.0s
=> CACHED [2/6] WORKDIR /usr/src/app 0.0s
=> CACHED [3/6] COPY go.mod ./ 0.0s
=> CACHED [4/6] RUN go mod download && go mod verify 0.0s
=> [5/6] COPY . . 0.0s
=> [6/6] RUN go build -v -o /usr/local/bin/app ./... 4.6s
=> exporting to image 0.1s
=> => exporting layers 0.1s
=> => writing image sha256:d546d176a115b935b73bc7a84c03dc03bf9d6e1c78d983fb3c71cc04faee9358 0.0s
=> => naming to docker.io/library/docker-cloud 0.0s
```

`docker run -dp 8090:80 docker-cloud`

```sh
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/docker-cloud$ docker run -dp 8090:80 docker-cloud
41da9ceefc966fabcea4da2b66aed409901dc992a3e93de670743bbecf7a5940
```

`curl localhost:8090`

```sh
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/docker-cloud$ curl localhost:8090
Hello, world.
```

`curl localhost:8090/ping`

```sh
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/docker-cloud$ curl localhost:8090/ping
pong
```

test the environment variable injection

`docker run -dp 8090:8080 -e HTTP_PORT=8080 docker-cloud`

```sh
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/docker-cloud$ docker run -dp 8090:8080 -e HTTP_PORT=8080 docker-cloud
d7fc97ec495ab4dc779e95bc5900509437cf7740b52c1a9022702d696c38b158
```

`curl localhost:8090`

```sh
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/docker-cloud$ curl localhost:8090
Hello, world.
```

`curl localhost:8090/ping`

```sh
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/docker-cloud$ curl localhost:8090/ping
pong
```

## Push to DockerHub

```sh
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/docker-cloud$ docker push bazmurphy/docker-cloud
Using default tag: latest
The push refers to repository [docker.io/bazmurphy/docker-cloud]
cb30e6886a9c: Pushed
46fc0466fe37: Pushed
1cb4fcd306f5: Layer already exists
9337039671d3: Pushed
180e03821175: Pushed
5f70bf18a086: Pushed
bae81d7f8189: Pushed
bf935cbb59a4: Pushed
01bd2df73a8f: Layer already exists
2353f7120e0e: Pushed
51a9318e6edf: Pushed
c5bb35826823: Pushed
latest: digest: sha256:3f11b513989ed6d1653945000d99bf42de94b693892a69f346fd8185824ee62c size: 2840
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/docker-cloud$
```
