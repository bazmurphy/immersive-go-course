# Bugs (or what I think are Bugs(?))

## Reported Bug 1

"An (imaginary) user of our app has reported that the note "#Monday Remember to take time for self-care" was behaving strangely... the tags didn't look right."

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

"Another user reported a bug where they deactivated their account, but were still able to see their notes. What's going on there?"

`/auth/auth.go`

There is no logic in the middleware to check if the user's status is "active" or "inactive" and to reject them if "inactive

So I can add some basic logic from Line 129+

```go
func (as *grpcAuthService) Verify(ctx context.Context, in *pb.VerifyRequest) (*pb.VerifyResponse, error) {
  ...
	// add logic to prevent "inactive" users from accessing any notes
	// log.Printf("DEBUG | id %s | password %s | status %s\n", row.id, row.password, row.status)
	if row.status == "inactive" {
		return &pb.VerifyResponse{
			State: pb.State_DENY,
		}, nil
	}
  ...
}
```

---

## Individual Note Route Error

`/api/api.go`

The individual `note` route is incorrect

In the README it says `/1/my/notes/:id.json`

But in the `Handler()` Line 136 it is `/1/my/note/`

```go
func (as *Service) Handler() http.Handler {
	mux := new(http.ServeMux)
	mux.HandleFunc("/1/my/note/", as.wrapAuth(as.authClient, as.handleMyNoteById))
	mux.HandleFunc("/1/my/notes.json", as.wrapAuth(as.authClient, as.handleMyNotes))
	return httplogger.HTTPLogger(mux)
}
```

So we need to add an `s` to `/1/my/note/`

```go
func (as *Service) Handler() http.Handler {
	mux := new(http.ServeMux)
	// [BUG]
	// mux.HandleFunc("/1/my/note/", as.wrapAuth(as.authClient, as.handleMyNoteById))
	mux.HandleFunc("/1/my/notes/", as.wrapAuth(as.authClient, as.handleMyNoteById))
	mux.HandleFunc("/1/my/notes.json", as.wrapAuth(as.authClient, as.handleMyNotes))
	return httplogger.HTTPLogger(mux)
}
```

And now it is resolved.

---

## User getting all their notes would query all notes not just the owners

`api/model/notes.go`

```go
	// [BUG]
	// queryRows, err := conn.Query(ctx, "SELECT id, owner, content, created, modified FROM public.note")
	queryRows, err := conn.Query(ctx, "SELECT id, owner, content, created, modified FROM public.note WHERE owner = $1", owner)
```

Now it will only query that specific owner's notes and not everyones

---

## Users can see other Users notes by ID

`api/api.go`

Before User1 could make a request by ID and see User2 note (if they knew the ID)

```go
	...
	requestUser, ok := authuserctx.FromAuthenticatedContext(ctx)
	...
		// if we get a note back but it is not owned by the request user then reject as unauthorized
	if err == nil && note.Owner != requestUser {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	...
```

I did ALSO consider moving this to the Model and adding another parameter (owner) and making the query more specific with a WHERE AND clause, and handling it there...

### User1 trying to query User2 note1 by ID

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/lxUr6TWQ.json -H 'Authorization: Basic bmNGV2JMWms6YmFuYW5h' -i
HTTP/1.1 401 Unauthorized
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Thu, 16 May 2024 14:38:53 GMT
Content-Length: 13

Unauthorized
baz@baz-pc:/buggy-app$
```

Now User1 is Unauthorized to query User2's note by ID

### User2 trying to query User2 note1 by ID

```sh
baz@baz-pc:/buggy-app$ curl 127.0.0.1:8090/1/my/notes/lxUr6TWQ.json -H 'Authorization: Basic M05qcVcxeHg6YXBwbGU=' -i
HTTP/1.1 200 OK
Content-Type: text/json
Date: Thu, 16 May 2024 14:39:25 GMT
Content-Length: 172

{"note":{"id":"lxUr6TWQ","owner":"3NjqW1xx","content":"user2 note1 with 0 tags","created":"2024-05-16T11:56:43.144587Z","modified":"2024-05-16T11:56:43.144587Z","tags":[]}}baz@baz-pc:/buggy-app$
```

But User2 can still query their own note by ID

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

---

## More Hints

1. Try out as a User: Work out if when logged in as a one user you can see the notes of the other user

2. Someone was suspicious that someone else had gotten their password

   - they saw changes to the their notes that they hand not made
   - they changed their password and they are still seeing changes to their own notes that they haven't made

3. Expect the Server to give us a status code in the 200-300-400 range

   - Can you make the server return a 500 code

4. Can you find ways the server could be faster

5. `http.Error` implicit / explicit returns ?

6. In terms of the Cache on the API Service, what problems may you have if you have multiple instances of these API Services, what does it make harder?
