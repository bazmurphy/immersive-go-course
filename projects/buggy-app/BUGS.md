# Bugs (or what I think are Bugs(?))

--

## Reported Bug 1

`/api/model/notes.go` Line 80

The regular expression is not working properly for the specific Bug Report Case

```go
func extractTags(input string) []string {
	re := regexp.MustCompile(`#([^#]+)`)
	matches := re.FindAllStringSubmatch(input, -1)
	tags := make([]string, 0, len(matches))
	for _, f := range matches {
		tags = append(tags, strings.TrimSpace(f[1]))
	}
	return tags
}
```

Make a new test for the reported bug

`/api/model/notes_test.go`

```go
// newly added test to address bug report
func TestTagsBugReport(t *testing.T) {
	text := "#Monday Remember to take time for self-care"
	expected := []string{"Monday"}

	tags := extractTags(text)

	if !reflect.DeepEqual(expected, tags) {
		t.Fatalf("expected %v, got %v", expected, tags)
	}
}
```

And it fails

So now fix the Regular Expression

```go
func extractTags(input string) []string {
	// re := regexp.MustCompile(`#([^#]+)`)
	// new regular expression to fix reported bug
	re := regexp.MustCompile(`#([\w]+)`)
	matches := re.FindAllStringSubmatch(input, -1)
	tags := make([]string, 0, len(matches))
	for _, f := range matches {
		tags = append(tags, strings.TrimSpace(f[1]))
	}
	return tags
}
```

Now the reported bug test passes

--

## Reported Bug 2

Run: `make migrate`

Then Run: `make run build`

Connect to the postgresql docker container
Then connect to postgresql and specifically the `app` database
Then query the `user` table to see the users and their `id` and `password` and `status`

```sh
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/buggy-app$ docker exec -it buggy-app-postgres-1 bash
root@03b211ceff98:/# psql -U postgres -d app
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
 usIgrmzp | $2y$10$O8VPlcAPa/iKHrkdyzN1cu7TvF5Goq6nRjSdaz9uXm1zPcVgRxQnK | 2024-05-14 17:48:42.571513 | 2024-05-14 17:48:42.579355 | inactive
 jBfa2Ww0 | $2y$10$wd5QGX9NNg6Kz1EKqn5pn.Ee6tiLem0pmjqF.tVeSPPsmG9PW9vUW | 2024-05-14 17:48:42.571513 | 2024-05-14 17:48:42.579355 | active
(2 rows)

app=#
```

The first user is `usIgrmzp` is `inactive`
The second user is `jBfa2Ww0` is `active`

From `/migrations/app/000002_create_dummy_users.up.sql`
We can see
the first user has `password: banana`
the second user has `password: apple`

So let's convert their user:password combinations into base64

`usIgrmzp:banana`

`jBfa2Ww0:apple`

```sh
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/buggy-app$ echo -n "usIgrmzp:banana" | base64
dXNJZ3JtenA6YmFuYW5h
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/buggy-app$ echo -n "jBfa2Ww0:apple" | base64
akJmYTJXdzA6YXBwbGU=
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/buggy-app$
```

Now we can use these to send the `curl` requests

```sh
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/buggy-app$ curl 127.0.0.1:8090/1/my/notes.json -H 'Authorization: Basic akJmYTJXdzA6YXBwbGU=' -i
HTTP/1.1 200 OK
Content-Type: text/json
Date: Tue, 14 May 2024 18:04:13 GMT
Content-Length: 12

{"notes":[]}baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/buggy-app$
```

```sh
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projectscurl 127.0.0.1:8090/1/my/notes.json -H 'Authorization: Basic dXNJZ3JtenA6YmFuYW5h' -i
HTTP/1.1 200 OK
Content-Type: text/json
Date: Tue, 14 May 2024 18:04:44 GMT
Content-Length: 12

{"notes":[]}baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/buggy-app$
```

And check the logs

```sh
api-1       | 2024/05/14 18:04:13 HTTP - 172.18.0.1:49128 - - 14/May/2024:18:04:13 +0000 "GET /1/my/notes.json HTTP/1.1" 200 12 curl/8.5.0 55387753us
api-1       | 2024/05/14 18:04:44 HTTP - 172.18.0.1:35190 - - 14/May/2024:18:04:44 +0000 "GET /1/my/notes.json HTTP/1.1" 200 12 curl/8.5.0 996173us
```

(BUG) `inactive` users should NOT be able to access their notes !!
So I need to check the authorization logic

---

NOTE:
(!!!) The database is going weird and I have to remake the users :S
Answer: This is probably because it is stored in `/tmp/` which gets wiped after reboot

```sh
app=# SELECT * FROM "user";
    id    |                           password                           |          created           |         modified          |  status
----------+--------------------------------------------------------------+----------------------------+---------------------------+----------
 gEh2w0__ | $2y$10$O8VPlcAPa/iKHrkdyzN1cu7TvF5Goq6nRjSdaz9uXm1zPcVgRxQnK | 2024-05-15 10:03:09.911299 | 2024-05-15 10:03:09.91911 | inactive
 IVm4D594 | $2y$10$wd5QGX9NNg6Kz1EKqn5pn.Ee6tiLem0pmjqF.tVeSPPsmG9PW9vUW | 2024-05-15 10:03:09.911299 | 2024-05-15 10:03:09.91911 | active
```

user 1 basic auth `Z0VoMncwX186YmFuYW5h`
user 2 basic auth `SVZtNEQ1OTQ6YXBwbGU=`

request from user 1 `curl 127.0.0.1:8090/1/my/notes.json -H 'Authorization: Basic Z0VoMncwX186YmFuYW5h' -i`
request from user 1 `curl 127.0.0.1:8090/1/my/notes.json -H 'Authorization: Basic SVZtNEQ1OTQ6YXBwbGU=' -i`

---

## JSON Indentation

`/api/api.go` Line 72

```go
  res, err := util.MarshalWithIndent(response, "")
```

No `ident` string value is ever passed into the `MarshalWithIndent` ?

We need to get this from the client somehow

It expects a `string` (optimally `"1"` to `"10"` that it can convert to an integer)

`/util/service.go` Line 17

```go
  if i, err := strconv.Atoi(indent); err == nil && i > 0 && i <= 10 {
    b, marshalErr = json.MarshalIndent(data, "", strings.Repeat(" ", i))
  } else {
    b, marshalErr = json.Marshal(data)
  }
```

There is no error handling of the `strconv.Itoa()` only a `err == nil`

(Not really a bug but `data interface{}` could be updated with `data any` (syntactic sugar))

--

## Using the outdated Context package address

The Go version is 1.19 by that time the std library `context` package should have been available?

So why are we importing `"golang.org/x/net/context"` everywhere

We should update this to `"context"`

---

## Inconsistent use of the Logger in the API Service

`/api/api.go` Line 31

```go
type Config struct {
  Port           int
  Log            *log.Logger // <--- here
  AuthServiceUrl string
  DatabaseUrl    string
}
```

There is very specifically a logger defined in the `Config` when you create a `New` `API Service`

But this logger is not consistently used...

There is a mix of both `as.config.Log.Printf()` and `fmt.Printf()` used... Why?

```go
  as.config.Log.Printf("api: route handler reached with invalid auth context")
  ...
  fmt.Printf("api: GetNotesForOwner failed: %v\n", err)
  ...
  fmt.Printf("api: response marshal failed: %v\n", err)
  ...
  fmt.Printf("api: no ID supplied: url path %v\n", r.URL.Path)
  ...
  fmt.Printf("api: GetNoteById failed: %v\n", err)
  ...
  fmt.Printf("api: response marshal failed: %v\n", err)
  ...
```

Whereas if you look at:

`/auth/auth.go` Line 20

```go
type Config struct {
  Port        int
  DatabaseUrl string
  Log         *log.Logger
}
```

There is also a logger defined in the `Config` when you create a `New` `Auth Service` but that seems to be consistently used

Another thought: Is it OK to be logging these Auth attempts... is this not a possible security issue?

---

## Incorrect Content-Type

`/api/api.go`

Line 79 inside `handleMyNotes`
Line 123 inside `handleNoteById`

Both have:

```go
  w.Header().Add("Content-Type", "text/json")
```

As I pointed out in my PRs to fix this on the course READMEs

`text/json` is not the correct MIME type

So we should use `application/json`

---

## API Shutdown

`/api/api.go` Line 172

```go
  server.Shutdown(context.TODO())
```

This feels off but I am not experienced enough with `context` to know if this right or not??

I thought `context.TODO()` is a placeholder

Should we be using the `ctx` that is passed into `Run()` ?

`/cmd/api/main_cmd.go`

line 29

```go
  ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
```

and passed in to `Run` Line 38

```go
  if err := as.Run(ctx); err != nil {
    log.Fatal(err)
  }
```

But the `ctx` that get's passed in gets cancelled with the `<--ctx.Done()` right ??

So maybe this is not a Bug and this is ok?

Because `Shutdown()` needs a context... and so an empty one is put in there??

But shouldn't it be `context.Background()` ??

---

## `Makefile`

Line 57

```makefile
run-database:
	docker compose up
```

This seems weird... this is bringing up the api & auth as well not just the Database?

I checked the profiles in `docker-compose.yml`

Shouldn't it be this:

```makefile
run-database:
    docker compose up postgres
```

---

## `docker-compose.yml`

### ports

Line 47 `auth` has `ports` `127.0.0.1:80:8080`

Line 63 `api` has `ports` `127.0.0.1:80:8090`

Why can't we just use `80:8080` `80:8090`. What is the necessity to bind to localhost only?

### order

Line 88-89

```yml
command: go test /app/...
environment:
  - POSTGRES_PASSWORD_FILE=/run/secrets/postgres-passwd
```

Shouldn't the `environment` be before the `command` like the others above?

---

## `.dockerignore` file

We could use a `.dockerignore` file to make sure anything we don't want is left out of the image

---
