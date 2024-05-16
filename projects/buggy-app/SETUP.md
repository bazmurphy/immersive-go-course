# Setup the App

## Create the Postgres Password

1. make a directory `volumes/secrets`
2. created a random password and put it in `volumes/secrets/postgres-passwd`

```Makefile
volumes/secrets/postgres-passwd:
	mkdir -p volumes/secrets
	# Create a random password for Postgres
	openssl rand -hex 24 | tr -d '\n' > volumes/secrets/postgres-passwd
```

## Create the Volume

This relies on `volumes/secrets/postgres-passwd`

1. make a directory `/home/baz/buggy-app-data`

```Makefile
volumes: volumes/secrets/postgres-passwd
	mkdir -p /home/baz/buggy-app-data
```

## Migrate the Database

This relies on `volumes`

1. run docker compose with the profiles `migrate build`
2. run docker compose `migrate` service

```Makefile
migrate: volumes
	docker compose --profile migrate build
	docker compose run migrate
```

```sh
baz@baz-pc:/buggy-app$ sudo make migrate
mkdir -p /home/baz/buggy-app-data
docker compose --profile migrate build
WARN[0000] /buggy-app/docker-compose.yml: `version` is obsolete
[+] Building 1.9s (41/53)                                                                                                                                                                                                                                                                                            docker:default
 => [migrate internal] load build definition from Dockerfile                                                                                                                                                                                                                                                                   0.0s
 => => transferring dockerfile: 752B                                                                                                                                                                                                                                                                                           0.0s
 => [api] resolve image config for docker-image://docker.io/docker/dockerfile:1                                                                                                                                                                                                                                                0.8s
 => [auth internal] load build definition from Dockerfile                                                                                                                                                                                                                                                                      0.0s
 => => transferring dockerfile: 752B                                                                                                                                                                                                                                                                                           0.0s
 => CACHED [api] docker-image://docker.io/docker/dockerfile:1@sha256:a57df69d0ea827fb7266491f2813635de6f17269be881f696fbfdf2d83dda33e                                                                                                                                                                                          0.0s
 => [api internal] load metadata for docker.io/library/golang:1.19-bullseye                                                                                                                                                                                                                                                    0.8s
 => [auth internal] load .dockerignore                                                                                                                                                                                                                                                                                         0.0s
 => => transferring context: 2B                                                                                                                                                                                                                                                                                                0.0s
 => [migrate internal] load .dockerignore                                                                                                                                                                                                                                                                                      0.0s
 => => transferring context: 2B                                                                                                                                                                                                                                                                                                0.0s
 => [migrate internal] load build context                                                                                                                                                                                                                                                                                      0.0s
 => => transferring context: 6.30kB                                                                                                                                                                                                                                                                                            0.0s
 => [api  1/14] FROM docker.io/library/golang:1.19-bullseye@sha256:2fdfcb03b1445f06f1cf8a342516bfd34026b527fef8427f40ea7b140168fda2                                                                                                                                                                                            0.0s
 => [auth internal] load build context                                                                                                                                                                                                                                                                                         0.0s
 => => transferring context: 6.30kB                                                                                                                                                                                                                                                                                            0.0s
 => CACHED [api  2/14] WORKDIR /app                                                                                                                                                                                                                                                                                            0.0s
 => CACHED [migrate  3/14] COPY go.mod ./                                                                                                                                                                                                                                                                                      0.0s
 => CACHED [migrate  4/14] COPY go.sum ./                                                                                                                                                                                                                                                                                      0.0s
 => CACHED [migrate  5/14] RUN go mod download                                                                                                                                                                                                                                                                                 0.0s
 => CACHED [migrate  6/14] COPY api ./api                                                                                                                                                                                                                                                                                      0.0s
 => CACHED [migrate  7/14] COPY auth ./auth                                                                                                                                                                                                                                                                                    0.0s
 => CACHED [migrate  8/14] COPY cmd ./cmd                                                                                                                                                                                                                                                                                      0.0s
 => CACHED [migrate  9/14] COPY migrations ./migrations                                                                                                                                                                                                                                                                        0.0s
 => CACHED [migrate 10/14] COPY util ./util                                                                                                                                                                                                                                                                                    0.0s
 => CACHED [migrate 11/14] RUN mkdir -p /out                                                                                                                                                                                                                                                                                   0.0s
 => CACHED [migrate 12/14] RUN go build -o /out ./...                                                                                                                                                                                                                                                                          0.0s
 => CACHED [migrate 13/14] COPY bin /bin                                                                                                                                                                                                                                                                                       0.0s
 => CACHED [auth 14/14] COPY migrations /migrations                                                                                                                                                                                                                                                                            0.0s
 => [migrate] exporting to image                                                                                                                                                                                                                                                                                               0.0s
 => => exporting layers                                                                                                                                                                                                                                                                                                        0.0s
 => => writing image sha256:7fca49e6ab8b83127f8483ed140609d19172762728d6777830766ee714af0c2f                                                                                                                                                                                                                                   0.0s
 => => naming to docker.io/library/buggy-app-migrate                                                                                                                                                                                                                                                                           0.0s
 => [auth] exporting to image                                                                                                                                                                                                                                                                                                  0.0s
 => => exporting layers                                                                                                                                                                                                                                                                                                        0.0s
 => => writing image sha256:5ae6f8b090c9f21ab1304ea265878711586bd5294bcba8c35e4d819f77dd8c5d                                                                                                                                                                                                                                   0.0s
 => => naming to docker.io/library/buggy-app-auth                                                                                                                                                                                                                                                                              0.0s
 => [api internal] load build definition from Dockerfile                                                                                                                                                                                                                                                                       0.0s
 => => transferring dockerfile: 752B                                                                                                                                                                                                                                                                                           0.0s
 => [api internal] load .dockerignore                                                                                                                                                                                                                                                                                          0.0s
 => => transferring context: 2B                                                                                                                                                                                                                                                                                                0.0s
 => [api internal] load build context                                                                                                                                                                                                                                                                                          0.0s
 => => transferring context: 1.98kB                                                                                                                                                                                                                                                                                            0.0s
 => CACHED [api  3/14] COPY go.mod ./                                                                                                                                                                                                                                                                                          0.0s
 => CACHED [api  4/14] COPY go.sum ./                                                                                                                                                                                                                                                                                          0.0s
 => CACHED [api  5/14] RUN go mod download                                                                                                                                                                                                                                                                                     0.0s
 => CACHED [api  6/14] COPY api ./api                                                                                                                                                                                                                                                                                          0.0s
 => CACHED [api  7/14] COPY auth ./auth                                                                                                                                                                                                                                                                                        0.0s
 => CACHED [api  8/14] COPY cmd ./cmd                                                                                                                                                                                                                                                                                          0.0s
 => CACHED [api  9/14] COPY migrations ./migrations                                                                                                                                                                                                                                                                            0.0s
 => CACHED [api 10/14] COPY util ./util                                                                                                                                                                                                                                                                                        0.0s
 => CACHED [api 11/14] RUN mkdir -p /out                                                                                                                                                                                                                                                                                       0.0s
 => CACHED [api 12/14] RUN go build -o /out ./...                                                                                                                                                                                                                                                                              0.0s
 => CACHED [api 13/14] COPY bin /bin                                                                                                                                                                                                                                                                                           0.0s
 => CACHED [api 14/14] COPY migrations /migrations                                                                                                                                                                                                                                                                             0.0s
 => [api] exporting to image                                                                                                                                                                                                                                                                                                   0.0s
 => => exporting layers                                                                                                                                                                                                                                                                                                        0.0s
 => => writing image sha256:8910c1a5b47c2c83c1542a6cb12922e220aa857d2c78b2708b90a5660ef8ff34                                                                                                                                                                                                                                   0.0s
 => => naming to docker.io/library/buggy-app-api                                                                                                                                                                                                                                                                               0.0s
docker compose run migrate
WARN[0000] /buggy-app/docker-compose.yml: `version` is obsolete
[+] Creating 1/0
 ✔ Container buggy-app-postgres-1  Created                                                                                                                                                                                                                                                                                     0.0s
[+] Running 1/1
 ✔ Container buggy-app-postgres-1  Started                                                                                                                                                                                                                                                                                     0.2s
wait-for-it.sh: waiting 60 seconds for postgres:5432
wait-for-it.sh: postgres:5432 is available after 1 seconds
2024/05/16 11:35:18 migrate: "file:///migrations/app" into "app" database
2024/05/16 11:35:18 migrate: complete
baz@baz-pc:/buggy-app$
```

---

## Build the App

This relies on `volumes`

1. build the go code
2. run docker compose with the profile `run build`

```Makefile
build: volumes
	@# If it doesn't build, we want to know ASAP
	go build ./...
	docker compose --profile run build
```

```sh
baz@baz-pc:/buggy-app$ make build
mkdir -p /home/baz/buggy-app-data
go build ./...
docker compose --profile run build
WARN[0000] /buggy-app/docker-compose.yml: `version` is obsolete
[+] Building 1.2s (37/37) FINISHED                                                                                                                                                                                                                                      docker:default
 => [auth internal] load build definition from Dockerfile                                                                                                                                                                                                                         0.0s
 => => transferring dockerfile: 752B                                                                                                                                                                                                                                              0.0s
 => [api] resolve image config for docker-image://docker.io/docker/dockerfile:1                                                                                                                                                                                                   0.5s
 => CACHED [api] docker-image://docker.io/docker/dockerfile:1@sha256:a57df69d0ea827fb7266491f2813635de6f17269be881f696fbfdf2d83dda33e                                                                                                                                             0.0s
 => [api internal] load metadata for docker.io/library/golang:1.19-bullseye                                                                                                                                                                                                       0.5s
 => [auth internal] load .dockerignore                                                                                                                                                                                                                                            0.0s
 => => transferring context: 2B                                                                                                                                                                                                                                                   0.0s
 => [api  1/14] FROM docker.io/library/golang:1.19-bullseye@sha256:2fdfcb03b1445f06f1cf8a342516bfd34026b527fef8427f40ea7b140168fda2                                                                                                                                               0.0s
 => [auth internal] load build context                                                                                                                                                                                                                                            0.0s
 => => transferring context: 1.98kB                                                                                                                                                                                                                                               0.0s
 => CACHED [api  2/14] WORKDIR /app                                                                                                                                                                                                                                               0.0s
 => CACHED [auth  3/14] COPY go.mod ./                                                                                                                                                                                                                                            0.0s
 => CACHED [auth  4/14] COPY go.sum ./                                                                                                                                                                                                                                            0.0s
 => CACHED [auth  5/14] RUN go mod download                                                                                                                                                                                                                                       0.0s
 => CACHED [auth  6/14] COPY api ./api                                                                                                                                                                                                                                            0.0s
 => CACHED [auth  7/14] COPY auth ./auth                                                                                                                                                                                                                                          0.0s
 => CACHED [auth  8/14] COPY cmd ./cmd                                                                                                                                                                                                                                            0.0s
 => CACHED [auth  9/14] COPY migrations ./migrations                                                                                                                                                                                                                              0.0s
 => CACHED [auth 10/14] COPY util ./util                                                                                                                                                                                                                                          0.0s
 => CACHED [auth 11/14] RUN mkdir -p /out                                                                                                                                                                                                                                         0.0s
 => CACHED [auth 12/14] RUN go build -o /out ./...                                                                                                                                                                                                                                0.0s
 => CACHED [auth 13/14] COPY bin /bin                                                                                                                                                                                                                                             0.0s
 => CACHED [auth 14/14] COPY migrations /migrations                                                                                                                                                                                                                               0.0s
 => [auth] exporting to image                                                                                                                                                                                                                                                     0.0s
 => => exporting layers                                                                                                                                                                                                                                                           0.0s
 => => writing image sha256:5ae6f8b090c9f21ab1304ea265878711586bd5294bcba8c35e4d819f77dd8c5d                                                                                                                                                                                      0.0s
 => => naming to docker.io/library/buggy-app-auth                                                                                                                                                                                                                                 0.0s
 => [api internal] load build definition from Dockerfile                                                                                                                                                                                                                          0.0s
 => => transferring dockerfile: 752B                                                                                                                                                                                                                                              0.0s
 => [api internal] load .dockerignore                                                                                                                                                                                                                                             0.0s
 => => transferring context: 2B                                                                                                                                                                                                                                                   0.0s
 => [api internal] load build context                                                                                                                                                                                                                                             0.0s
 => => transferring context: 1.98kB                                                                                                                                                                                                                                               0.0s
 => CACHED [api  3/14] COPY go.mod ./                                                                                                                                                                                                                                             0.0s
 => CACHED [api  4/14] COPY go.sum ./                                                                                                                                                                                                                                             0.0s
 => CACHED [api  5/14] RUN go mod download                                                                                                                                                                                                                                        0.0s
 => CACHED [api  6/14] COPY api ./api                                                                                                                                                                                                                                             0.0s
 => CACHED [api  7/14] COPY auth ./auth                                                                                                                                                                                                                                           0.0s
 => CACHED [api  8/14] COPY cmd ./cmd                                                                                                                                                                                                                                             0.0s
 => CACHED [api  9/14] COPY migrations ./migrations                                                                                                                                                                                                                               0.0s
 => CACHED [api 10/14] COPY util ./util                                                                                                                                                                                                                                           0.0s
 => CACHED [api 11/14] RUN mkdir -p /out                                                                                                                                                                                                                                          0.0s
 => CACHED [api 12/14] RUN go build -o /out ./...                                                                                                                                                                                                                                 0.0s
 => CACHED [api 13/14] COPY bin /bin                                                                                                                                                                                                                                              0.0s
 => CACHED [api 14/14] COPY migrations /migrations                                                                                                                                                                                                                                0.0s
 => [api] exporting to image                                                                                                                                                                                                                                                      0.0s
 => => exporting layers                                                                                                                                                                                                                                                           0.0s
 => => writing image sha256:8910c1a5b47c2c83c1542a6cb12922e220aa857d2c78b2708b90a5660ef8ff34                                                                                                                                                                                      0.0s
 => => naming to docker.io/library/buggy-app-api                                                                                                                                                                                                                                  0.0s
baz@baz-pc:/buggy-app$
```

---

## Run the App

1. run docker compose with the profile `run`

```Makefile
run:
	docker compose --profile run up
```

```sh
baz@baz-pc:/buggy-app$ make run
docker compose --profile run up
WARN[0000] /buggy-app/docker-compose.yml: `version` is obsolete
[+] Running 3/0
 ✔ Container buggy-app-postgres-1  Running                                                                                                                                                                                                                                        0.0s
 ✔ Container buggy-app-auth-1      Created                                                                                                                                                                                                                                        0.0s
 ✔ Container buggy-app-api-1       Created                                                                                                                                                                                                                                        0.0s
Attaching to api-1, auth-1, postgres-1
auth-1      | wait-for-it.sh: waiting 60 seconds for postgres:5432
auth-1      | wait-for-it.sh: postgres:5432 is available after 0 seconds
auth-1      | 2024/05/16 11:39:25 auth service: listening: :80
api-1       | wait-for-it.sh: waiting 60 seconds for postgres:5432
api-1       | wait-for-it.sh: postgres:5432 is available after 0 seconds
api-1       | 2024/05/16 11:39:25 api service: listening: :80
```

---

## Find out the two user IDs

```sh
baz@baz-pc:/buggy-app$ docker exec -it buggy-app-postgres-1 bash
root@ad9b1da1b77f:/# psql -U postgres -d app
psql (16.3 (Debian 16.3-1.pgdg120+1))
Type "help" for help.

app=# \dt
               List of relations
 Schema |       Name        | Type  |  Owner
--------+-------------------+-------+----------
 public | note              | table | postgres
 public | schema_migrations | table | postgres
 public | user              | table | postgres
(3 rows)

app=# SELECT * FROM "user";
    id    |                           password                           |          created           |          modified          |  status
----------+--------------------------------------------------------------+----------------------------+----------------------------+----------
 ncFWbLZk | $2y$10$O8VPlcAPa/iKHrkdyzN1cu7TvF5Goq6nRjSdaz9uXm1zPcVgRxQnK | 2024-05-16 11:35:18.365412 | 2024-05-16 11:35:18.373774 | inactive
 3NjqW1xx | $2y$10$wd5QGX9NNg6Kz1EKqn5pn.Ee6tiLem0pmjqF.tVeSPPsmG9PW9vUW | 2024-05-16 11:35:18.365412 | 2024-05-16 11:35:18.373774 | active
(2 rows)

app=#
```

User1 ID: `ncFWbLZk`

User2 ID: `3NjqW1xx`

---

## Generate the Basic Authentication for the two users

User1 ID: `ncFWbLZk`

User1 Password: `banana`

User2 ID: `3NjqW1xx`

User2 Password: `apple`

Now generate the base64 id:password combination

```sh
baz@baz-pc:/buggy-app$ echo -n "ncFWbLZk:banana" | base64
bmNGV2JMWms6YmFuYW5h
baz@baz-pc:/buggy-app$ echo -n "3NjqW1xx:apple" | base64
M05qcVcxeHg6YXBwbGU=
baz@baz-pc:/buggy-app$
```

User1 Auth: `bmNGV2JMWms6YmFuYW5h`

User2 Auth: `M05qcVcxeHg6YXBwbGU=`

These can now be used for `curl` requests

---

## Make the users Notes

There are no notes to start with:

```sh
baz@baz-pc:/buggy-app$ docker exec -it buggy-app-postgres-1 bash
root@ad9b1da1b77f:/# psql -U postgres -d app
psql (16.3 (Debian 16.3-1.pgdg120+1))
Type "help" for help.

app=# \dt
               List of relations
 Schema |       Name        | Type  |  Owner
--------+-------------------+-------+----------
 public | note              | table | postgres
 public | schema_migrations | table | postgres
 public | user              | table | postgres
(3 rows)

app=# SELECT * FROM "note";
 id | owner | content | created | modified
----+-------+---------+---------+----------
(0 rows)

app=#
```

So let's use the `/cmd/migrate/migrate_cmd.go` CLI tool to create some

```sh
baz@baz-pc:/buggy-app$ go run ./cmd/test note -owner ncFWbLZk -content "user1 note1 with 2 tags #sometag1 #sometag2"
2024/05/16 12:56:20 new note created
2024/05/16 12:56:20     id: fLLJ1DeX
2024/05/16 12:56:20     owner: ncFWbLZk
2024/05/16 12:56:20     content: "user1 note1 with 2 tags #sometag1 #sometag2"
baz@baz-pc:/buggy-app$ go run ./cmd/test note -owner ncFWbLZk -content "user1 note2 with 1 tag #sometag1"
2024/05/16 12:56:26 new note created
2024/05/16 12:56:26     id: SI-brU7V
2024/05/16 12:56:26     owner: ncFWbLZk
2024/05/16 12:56:26     content: "user1 note2 with 1 tag #sometag1"
baz@baz-pc:/buggy-app$ go run ./cmd/test note -owner ncFWbLZk -content "user1 note3 with 0 tags"
2024/05/16 12:56:35 new note created
2024/05/16 12:56:35     id: S-opG2sL
2024/05/16 12:56:35     owner: ncFWbLZk
2024/05/16 12:56:35     content: "user1 note3 with 0 tags"
baz@baz-pc:/buggy-app$ go run ./cmd/test note -owner 3NjqW1xx -content "user2 note1 with 0 tags"
2024/05/16 12:56:43 new note created
2024/05/16 12:56:43     id: lxUr6TWQ
2024/05/16 12:56:43     owner: 3NjqW1xx
2024/05/16 12:56:43     content: "user2 note1 with 0 tags"
baz@baz-pc:/buggy-app$ go run ./cmd/test note -owner 3NjqW1xx -content "user2 note2 with 2 tags #anothertag1 #anothertag2"
2024/05/16 12:56:52 new note created
2024/05/16 12:56:52     id: bxuYPp0r
2024/05/16 12:56:52     owner: 3NjqW1xx
2024/05/16 12:56:52     content: "user2 note2 with 2 tags #anothertag1 #anothertag2"
baz@baz-pc:/buggy-app$ go run ./cmd/test note -owner 3NjqW1xx -content "user2 note3 with 1 tag #anothertag1"
2024/05/16 12:56:58 new note created
2024/05/16 12:56:58     id: PZU8GVDj
2024/05/16 12:56:58     owner: 3NjqW1xx
2024/05/16 12:56:58     content: "user2 note3 with 1 tag #anothertag1"
baz@baz-pc:/buggy-app$
```

Check the database to see if they were all created successfully

```sh
baz@baz-pc:/buggy-app$ docker exec -it buggy-app-postgres-1 bash
root@ad9b1da1b77f:/# psql -U postgres -d app
psql (16.3 (Debian 16.3-1.pgdg120+1))
Type "help" for help.

app=# \dt
               List of relations
 Schema |       Name        | Type  |  Owner
--------+-------------------+-------+----------
 public | note              | table | postgres
 public | schema_migrations | table | postgres
 public | user              | table | postgres
(3 rows)

app=# SELECT * FROM "note";
    id    |  owner   |                      content                      |          created           |          modified
----------+----------+---------------------------------------------------+----------------------------+----------------------------
 fLLJ1DeX | ncFWbLZk | user1 note1 with 2 tags #sometag1 #sometag2       | 2024-05-16 11:56:20.510991 | 2024-05-16 11:56:20.510991
 SI-brU7V | ncFWbLZk | user1 note2 with 1 tag #sometag1                  | 2024-05-16 11:56:26.537306 | 2024-05-16 11:56:26.537306
 S-opG2sL | ncFWbLZk | user1 note3 with 0 tags                           | 2024-05-16 11:56:35.532484 | 2024-05-16 11:56:35.532484
 lxUr6TWQ | 3NjqW1xx | user2 note1 with 0 tags                           | 2024-05-16 11:56:43.144587 | 2024-05-16 11:56:43.144587
 bxuYPp0r | 3NjqW1xx | user2 note2 with 2 tags #anothertag1 #anothertag2 | 2024-05-16 11:56:52.244208 | 2024-05-16 11:56:52.244208
 PZU8GVDj | 3NjqW1xx | user2 note3 with 1 tag #anothertag1               | 2024-05-16 11:56:58.230549 | 2024-05-16 11:56:58.230549
(6 rows)

app=#
```

User1 Note1 ID: `fLLJ1DeX`

User1 Note2 ID: `SI-brU7V`

User1 Note3 ID: `S-opG2sL`

User2 Note1 ID: `lxUr6TWQ`

User2 Note2 ID: `bxuYPp0r`

User2 Note3 ID: `PZU8GVDj`

## All Possible Requests

Now lets make a list of all relevant `curl` requests to test

### Endpoints

The available API endpoints are:

`/1/my/notes.json` - GET ALL notes owned by the authenticated user

`/1/my/notes/:id.json` - GET a specific note owned by the authenticated user

### User1 request for all their own notes

`curl 127.0.0.1:8090/1/my/notes.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i
HTTP/1.1 200 OK
Content-Type: text/json
Date: Thu, 16 May 2024 12:00:01 GMT
Content-Length: 563

{"notes":[{"id":"fLLJ1DeX","owner":"ncFWbLZk","content":"user1 note1 with 2 tags #sometag1 #sometag2","created":"2024-05-16T11:56:20.510991Z","modified":"2024-05-16T11:56:20.510991Z","tags":["sometag1","sometag2"]},{"id":"SI-brU7V","owner":"ncFWbLZk","content":"user1 note2 with 1 tag #sometag1","created":"2024-05-16T11:56:26.537306Z","modified":"2024-05-16T11:56:26.537306Z","tags":["sometag1"]},{"id":"S-opG2sL","owner":"ncFWbLZk","content":"user1 note3 with 0 tags","created":"2024-05-16T11:56:35.532484Z","modified":"2024-05-16T11:56:35.532484Z","tags":[]}]}
baz@baz-pc:/buggy-app$
```

### User2 request for all their own notes

`curl 127.0.0.1:8090/1/my/notes.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i
HTTP/1.1 200 OK
Content-Type: text/json
Date: Thu, 16 May 2024 12:09:28 GMT
Content-Length: 581

{"notes":[{"id":"lxUr6TWQ","owner":"3NjqW1xx","content":"user2 note1 with 0 tags","created":"2024-05-16T11:56:43.144587Z","modified":"2024-05-16T11:56:43.144587Z","tags":[]},{"id":"bxuYPp0r","owner":"3NjqW1xx","content":"user2 note2 with 2 tags #anothertag1 #anothertag2","created":"2024-05-16T11:56:52.244208Z","modified":"2024-05-16T11:56:52.244208Z","tags":["anothertag1","anothertag2"]},{"id":"PZU8GVDj","owner":"3NjqW1xx","content":"user2 note3 with 1 tag #anothertag1","created":"2024-05-16T11:56:58.230549Z","modified":"2024-05-16T11:56:58.230549Z","tags":["anothertag1"]}]}
baz@baz-pc:/buggy-app$
```

### User1 request for their note1 (note id: `fLLJ1DeX`)

`curl 127.0.0.1:8090/1/my/notes/fLLJ1DeX.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/fLLJ1DeX.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Thu, 16 May 2024 12:12:22 GMT
Content-Length: 19

404 page not found
```

(BUG) it should retrieve the note1 owned by user1

### User1 request for their note2 (note id: `SI-brU7V`)

`curl 127.0.0.1:8090/1/my/notes/SI-brU7V.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/SI-brU7V.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Thu, 16 May 2024 12:17:02 GMT
Content-Length: 19

404 page not found
```

(BUG) it should retrieve the note2 owned by user1

### User1 request for their note3 (note id: `S-opG2sL`)

`curl 127.0.0.1:8090/1/my/notes/S-opG2sL.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i`

```
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/S-opG2sL.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Thu, 16 May 2024 12:17:16 GMT
Content-Length: 19

404 page not found
```

(BUG) it should retrieve the note3 owned by user1

### User2 request for their note1 (note id: `lxUr6TWQ`)

`curl 127.0.0.1:8090/1/my/notes/lxUr6TWQ.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/lxUr6TWQ.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Thu, 16 May 2024 12:19:35 GMT
Content-Length: 19

404 page not found
```

(BUG) it should retrieve the note1 owned by user2

### User2 request for their note2 (note id: `bxuYPp0r`)

`curl 127.0.0.1:8090/1/my/notes/bxuYPp0r.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/bxuYPp0r.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Thu, 16 May 2024 12:20:06 GMT
Content-Length: 19

404 page not found
```

(BUG) it should retrieve the note2 owned by user2

### User2 request for their note3 (note id: `PZU8GVDj`)

`curl 127.0.0.1:8090/1/my/notes/PZU8GVDj.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/PZU8GVDj.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Thu, 16 May 2024 12:20:17 GMT
Content-Length: 19

404 page not found
```

(BUG) it should retrieve the note3 owned by user2

Fixed in `BUGS.md` `## Reported Bug 2`

---

## status `active` vs `inactive`

User1 `status` is `inactive`

BEFORE: (BUG) User1 can retrieve notes (as seen above) when they should not be able to

Fixed in `BUGS.md` `## Reported Bug 2`

AFTER:

### User1 `inactive` request for all their notes

`curl 127.0.0.1:8090/1/my/notes.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i
HTTP/1.1 401 Unauthorized
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Thu, 16 May 2024 13:13:10 GMT
Content-Length: 13

Unauthorized
baz@baz-pc:/buggy-app$
```

User1 is now Unauthorized to see all their notes

### User1 `inactive` request for their note1 (note id: `fLLJ1DeX`)

`curl 127.0.0.1:8090/1/my/notes/fLLJ1DeX.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/fLLJ1DeX.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i
HTTP/1.1 401 Unauthorized
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Thu, 16 May 2024 13:13:47 GMT
Content-Length: 13

Unauthorized
baz@baz-pc:/buggy-app$
```

User1 is now Unauthorized to see their note1

### User2 `active` request for all their notes

`curl 127.0.0.1:8090/1/my/notes.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i
HTTP/1.1 200 OK
Content-Type: text/json
Date: Thu, 16 May 2024 13:15:07 GMT
Content-Length: 581

{"notes":[{"id":"lxUr6TWQ","owner":"3NjqW1xx","content":"user2 note1 with 0 tags","created":"2024-05-16T11:56:43.144587Z","modified":"2024-05-16T11:56:43.144587Z","tags":[]},{"id":"bxuYPp0r","owner":"3NjqW1xx","content":"user2 note2 with 2 tags #anothertag1 #anothertag2","created":"2024-05-16T11:56:52.244208Z","modified":"2024-05-16T11:56:52.244208Z","tags":["anothertag1","anothertag2"]},{"id":"PZU8GVDj","owner":"3NjqW1xx","content":"user2 note3 with 1 tag #anothertag1","created":"2024-05-16T11:56:58.230549Z","modified":"2024-05-16T11:56:58.230549Z","tags":["anothertag1"]}]}
baz@baz-pc:/buggy-app$
```

User2 `active` is Authorized to see all their own notes

### User2 `active` request for their note1 (note id: `lxUr6TWQ`)

`curl 127.0.0.1:8090/1/my/notes/lxUr6TWQ.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/lxUr6TWQ.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i
HTTP/1.1 200 OK
Content-Type: text/json
Date: Thu, 16 May 2024 13:15:22 GMT
Content-Length: 172

{"note":{"id":"lxUr6TWQ","owner":"3NjqW1xx","content":"user2 note1 with 0 tags","created":"2024-05-16T11:56:43.144587Z","modified":"2024-05-16T11:56:43.144587Z","tags":[]}}
baz@baz-pc:/buggy-app$
```

User2 `active` is Authorized to see their note1

---

(!!!) After fixing the `/note` to `/notes` bug above - AND - adjusting User1 to `active` from formerly `inactive`

Try to request notes which do NOT belong the authenticated user

### User1 request for User2 note1

`curl 127.0.0.1:8090/1/my/notes/lxUr6TWQ.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/lxUr6TWQ.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i
HTTP/1.1 200 OK
Content-Type: text/json
Date: Thu, 16 May 2024 12:46:32 GMT
Content-Length: 172

{"note":{"id":"lxUr6TWQ","owner":"3NjqW1xx","content":"user2 note1 with 0 tags","created":"2024-05-16T11:56:43.144587Z","modified":"2024-05-16T11:56:43.144587Z","tags":[]}}
```

(BUG) User1 can see User2 note1 - this should NOT be allowed

### User1 request for User2 note2

`curl 127.0.0.1:8090/1/my/notes/bxuYPp0r.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/bxuYPp0r.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i
HTTP/1.1 200 OK
Content-Type: text/json
Date: Thu, 16 May 2024 12:47:44 GMT
Content-Length: 225

{"note":{"id":"bxuYPp0r","owner":"3NjqW1xx","content":"user2 note2 with 2 tags #anothertag1 #anothertag2","created":"2024-05-16T11:56:52.244208Z","modified":"2024-05-16T11:56:52.244208Z","tags":["anothertag1","anothertag2"]}}baz@baz-pc:/buggy-app$
```

(BUG) User1 can see User2 note2 - this should NOT be allowed

### User1 request for User2 note3

`curl 127.0.0.1:8090/1/my/notes/PZU8GVDj.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/PZU8GVDj.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i
HTTP/1.1 200 OK
Content-Type: text/json
Date: Thu, 16 May 2024 12:48:44 GMT
Content-Length: 197

{"note":{"id":"PZU8GVDj","owner":"3NjqW1xx","content":"user2 note3 with 1 tag #anothertag1","created":"2024-05-16T11:56:58.230549Z","modified":"2024-05-16T11:56:58.230549Z","tags":["anothertag1"]}}
```

(BUG) User1 can see User2 note3 - this should NOT be allowed

### User2 request for User1 note1

`curl 127.0.0.1:8090/1/my/notes/fLLJ1DeX.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/fLLJ1DeX.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i
HTTP/1.1 200 OK
Content-Type: text/json
Date: Thu, 16 May 2024 12:50:48 GMT
Content-Length: 213

{"note":{"id":"fLLJ1DeX","owner":"ncFWbLZk","content":"user1 note1 with 2 tags #sometag1 #sometag2","created":"2024-05-16T11:56:20.510991Z","modified":"2024-05-16T11:56:20.510991Z","tags":["sometag1","sometag2"]}}
```

(BUG) User2 can see User1 note1 - this should NOT be allowed

### User2 request for User1 note2

`curl 127.0.0.1:8090/1/my/notes/SI-brU7V.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/SI-brU7V.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i
HTTP/1.1 200 OK
Content-Type: text/json
Date: Thu, 16 May 2024 12:51:29 GMT
Content-Length: 191

{"note":{"id":"SI-brU7V","owner":"ncFWbLZk","content":"user1 note2 with 1 tag #sometag1","created":"2024-05-16T11:56:26.537306Z","modified":"2024-05-16T11:56:26.537306Z","tags":["sometag1"]}}
```

(BUG) User2 can see User1 note2 - this should NOT be allowed

### User2 request for User1 note3

`curl 127.0.0.1:8090/1/my/notes/S-opG2sL.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i`

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/S-opG2sL.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i
HTTP/1.1 200 OK
Content-Type: text/json
Date: Thu, 16 May 2024 12:51:42 GMT
Content-Length: 172

{"note":{"id":"S-opG2sL","owner":"ncFWbLZk","content":"user1 note3 with 0 tags","created":"2024-05-16T11:56:35.532484Z","modified":"2024-05-16T11:56:35.532484Z","tags":[]}}
```

(BUG) User2 can see User1 note3 - this should NOT be allowed

---
