# Notes to Understand the App

---

## `/cmd/api/main_cmd.go`

I will start here as the entry point of the API

-get the port it will listen on

-read the postgres password from the env or a file

-setup a context that can be sent a signal to graceful shutdown

-make a new api service
passing it:
-port
-a logger
-a url "auth:80" ?? (weird?)
-a database url (using the password above) ("postgres://postgres:%s@postgres:5432/app", passwd)

-run the api service passing it the context

-if there is an error then fatal

---

## `/api/api.go`

### `DbClient` `interface`

-with 3 methods  
-QueryRow  
-Query  
-Close

### `Config` `struct`

-this is what was used above in `/cmd/api/main_cmd.go`
-Port  
-Log  
-AuthServiceUrl  
-DatabaseUrl

### `Service` `struct`

-config (Config)  
-authClient (auth.Client)  
-pool (DbClient)

### `New` Service constructor

returns a new Service  
and takes in  
-config  
-but where is the authClient (auth.Client) ??  
-where is the pool (DbClient) ??

### `Run()`

`func (as *Service) Run(ctx context.Context) error {}`

`listen` gets the port form the `config.Port`

pgsql `pool` created, passed:  
`context`
`as.config.AuthServiceUrl` (API Service config)

`as.pool` is SET at this point from the pool (created just above)

`auth.NewClient(ctx, as.config.AuthServiceUrl)`

`as.authClient` is SET at this point from the client (created just above)

`mux` is created by `as.Handler()`  
mux is an abbreviation of multiplex  
which is the root handler that works out where to route the various requests

`server` is created from a new `http.Server{}` taking  
`Addr: listen` (which is the `listen` port created above)  
`Handler: mux` (which is the `mux` created above)

`runErr` created

`wg` wait group created  
-add 1 to the wg
-spawn a new goroutine (but why do we do this in a new goroutine ??)  
`defer wg.Done()` -`server.ListenAndServe()` runs the server

`as.config.Log.Printf` we write a message to the logger

`<-ctx.Done()` if we receive the context Done

`server.Shutdown(context.TODO())` (is it ok to use .TODO() here ??)

`wg.Wait()` wait for the waitgroups to finish (the single goroutine that is running the server)  
`return runErr` if any

### `Handler()`

`func (as *Service) Handler() http.Handler {}`

`mux := new(http.ServeMux)`  
this is what determines how to handle the routes  
"`ServeMux` is an HTTP request multiplexer. It matches the URL of each incoming request against a list of registered patterns and calls the handler for the pattern that most closely matches the URL."

`mux.HandleFunc("/1/my/note/", as.wrapAuth(as.authClient, as.handleMyNoteById))`  
/1/my/note/ uses `handleMyNotById` handler

`mux.HandleFunc("/1/my/notes.json", as.wrapAuth(as.authClient, as.handleMyNotes))`
/1/my/notes.json uses `handleMyNotes` handler

both are wrapped in `as.wrapAuth` (which is like a middleware)

`return httplogger.HTTPLogger(mux)` why are we wrapping the `mux` in this ??

### `handleMyNotes()`

`func (as *Service) handleMyNotes(w http.ResponseWriter, r *http.Request) {}`

`ctx := r.Context()`  
-"Context returns the request's context."

`owner, ok := authuserctx.FromAuthenticatedContext(ctx)`  
-get the authenticated user from the context -- this will have been written earlier  
-if not ok return an http error

`notes, err := model.GetNotesForOwner(ctx, as.pool, owner)`  
-use the "model" layer to get a list of the owner's notes  
-if not ok return http error

-create a `response` struct (with the `model.Notes`)

`res, err := util.MarshalWithIndent(response, "")`  
-why is there an empty string here ?? shouldn't it be the indent amount ??  
-if not ok print error

`w.Header().Add("Content-Type", "text/json")`
-write the header
-this should be `application/json` !! [BUG]

`w.Write(res)`  
-write the body

Q: why are some of these logging to the logger and some just to Printf ??

### `handleMyNoteById()`

`func (as *Service) handleMyNoteById(w http.ResponseWriter, r *http.Request) {}`

`ctx := r.Context()`  
-"Context returns the request's context."

`_, ok := authuserctx.FromAuthenticatedContext(ctx)`  
-why this time do we not use `owner` ?? [BUG] (like the above)  
-if not ok return http error

`id := strings.Replace(path.Base(r.URL.Path), ".json", "", 1)`  
-get the id from the url path  
-is this process of stripping out the id correct ??  
-if no id then error

`note, err := model.GetNoteById(ctx, as.pool, id)`  
-use the "model" layer to get a list of the owner's notes  
-how do we get the owner here?
-and why are we allowing to get all the owner's notes? can everyone access this ?? look at the auth after this ??
-if err then error (failed to get the note)

`response := struct`  
-build a response with `model.Note`

`res, err := util.MarshalWithIndent(response, "")`  
-again we are passing an empty string into the `util.MarshalWithIdent`... why ?? [BUG] ??

`w.Header().Add("Content-Type", "text/json")`
-write the header
-this should be `application/json` !! [BUG]

-`w.Write(res)`  
-write the body

## `/api/api_auth.go`

### `wrapAuth()`

`func (as *Service) wrapAuth(client auth.Client, handler http.HandlerFunc) http.HandlerFunc {}`

`wrapAuth` takes a handler function (likely to be the API endpoint)  
and wraps it with an authentication check using an `AuthClient`
If the authentication passes, it adds the authenticated user ID to the context  
using the `authuserctx` package, and then calls the inner handler.  
The ID can be retrieved later using the `FromAuthenticatedContext` function.

-arguments:  
-`client` `auth.Client`  
-`handler` `http.HandlerFunc`

`return func(w http.ResponseWriter, r *http.Request) {}`
-it returns a function that is an `http.HandlerFunc` -`ctx, cancel := context.WithCancel(r.Context())`
-what is `r.Context()` ? ".Context() returns the request's context. "
-do we actually use the `context.WithCancel()`?

`id, passwd, ok := r.BasicAuth()`
-if not ok then return http error

`result, err := client.Verify(ctx, id, passwd)`  
-use the auth client to check if this id/password combo is approved  
-`client.Verify()` is coming from `/auth/client.go`  
-if err then return http error

`if result.State != auth.StateAllow`  
-unless we get an allow then deny  
-`.State` is coming from `/auth/client.go`  
-`auth.StateAllow` is coming from `/auth/client.go`

`ctx = authuserctx.NewAuthenticatedContext(ctx, id)`
-add the ID to the context and call the inner handler  
-`NewAuthenticatedContext` is coming from `/util/authuserctx`

`handler(w, r.WithContext(ctx))`
-returns a reader with `ctx` from the line above^
-returns a `handler` - i assume this is an implicit return?

## `/api/model/notes.go`

### `type` `Note` `struct`

-`Id` `string` `json:"id"`  
-`Owner` `string` `json:"owner"`  
-`Content` `string` `json:"content"`  
-`Created` `time.Time` `json:"created"`  
-`Modified` `time.Time` `json:"modified"`  
-`Tags` `[]string` `json:"tags"`
-the type for a Note

### `type` `Notes` `[]Note`

-the type for Notes (a slice of many `Note`)

### `type dbConn interface {}`

-the interface for dbConn  
-with two methods:  
-`Query()`  
-`QueryRow`

### `func GetNotesForOwner(ctx context.Context, conn dbConn, owner string) (Notes, error) {}`

-arguments:  
-`ctx` `context.Context`  
-`conn` `dbConn`  
-`owner` `string`

`if owner == ""`  
-the query needs an owner

`queryRows, err := conn.Query(ctx, "SELECT id, owner, content, created, modified FROM public.note")`  
-select id|owner|content|created from `public.note`  
-`Next()` through the rows and populate a `note` and append to `notes`  
-why are we not querying directly for that specific owner id ?? [BUG] ??  
-`if note.Owner == owner`  
-why are we doing this check? we only get the tags if it is our OWN notes ??

-if there is an error going through the rows then return

-return `notes, nil`

### `func GetNoteById(ctx context.Context, conn dbConn, id string) (Note, error) {}`

-arguments:  
-`ctx` `context.Context`  
-`conn` `dbConn`  
-`id` `string`

`if id == ""`  
-the query needs an id
-return an empty note ?? why are we doing this ?? we should return `nil` ?? [BUG]

`row := conn.QueryRow(ctx, "SELECT id, owner, content, created, modified FROM public.note WHERE id = $1", id)`  
-select id|owner|content|created|modified from `public.note` where id = `id`

`err := row.Scan(&note.Id, &note.Owner, &note.Content, &note.Created, &note.Modified)`  
-if err then return an empty note and the error ?? why are we doing this ?? we should return `nil` ?? [BUG]

`note.Tags = extractTags(note.Content)`  
-get the tags from the note

`return note, nil`  
-return the note and no error

### `func extractTags(input string) []string {}`

-arguments:  
-`input` `string`

`re := regexp.MustCompile('#([^#]+)')`  
-get every tag that starts with # and has any character after that which is not #  
-is this regex correct ?? investigate it  
`matches := re.FindAllStringSubmatch(input, -1)`  
-find all occurences of the pattern in the `input` string  
-`-1` means find all matches not just the first ??  
`tags := make([]string, 0, len(matches))`  
-make a string slice, length 0 and length of the matches found above  
`for _, f := range matches {}`  
-iterate over the `matches`  
`tags = append(tags, strings.TrimSpace(f[1]))`  
-append the tag with any leading/trailing whitespace removed  
-why do we need `f[1]` ?? to remove the `#` ??
`return tags`  
-return the tags slice

---

So that is all the `api` folder covered

Now let's move onto the `auth` folder

## `/cmd/auth/main_cmd.go`

`port := flag.Int("port", 80, "port the server will listen on")`  
-is the port meant to be `80`, the same as the `api` ??

`passwd, err := util.ReadPasswd()`  
-same logic as the `api` use the `util.ReadPasswd()` function

`ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)`  
-same logic as the `api` create a `NotifyContext`  
-what happens if `auth` recieves a shutdown signal but `api` does not ??

`as := auth.New(auth.Config{}`  
-create a new auth service  
-takes an `auth.Config` struct, with:  
`Port: *port`  
`DatabaseURL: fmt.Sprintf("postgres://postgres:%s@postgres:5432/app", passwd),`
`Log: log.Default()`
-why does the auth service need access to the database ??
-i assume we are creating users and storing their hashed/salted passwords there?

`if err := as.Run(ctx); err != nil {}`
-if we get an err from the `Run()` then fatal exit

## `/auth/auth.go`

`type Config struct`  
-`Port` `int`  
-`DatabaseUrl` `string`  
-`Log` `*log.Logger`  
-this is the auth service config

`type Service struct`  
-`config` `Config`  
-`grpcService` `*grpcAuthService`  
-this is the auth service

### `func New(config Config) *Service {}`

-the new service constructor

### `func (as *Service) Run(ctx context.Context) error {}`

-starts the underlying gRPC server according to the supplied `Config`

`pool, err := pgxpool.New(ctx, as.config.DatabaseUrl)`  
-connect to the database with the `config.DatabaseUrl`

`as.grpcService.pool = pool`  
-add the pool to the "inner" auth service which implements the gRPC interface and responds to RPCs

`listen := fmt.Sprintf(":%d", as.config.Port)`  
-create a TCP listener for the gRPC server to use

`lis, err := net.Listen("tcp", listen)`  
-"Listen() announces on the local network address."

`grpcServer := grpc.NewServer()`  
-set up and register the server

`pb.RegisterAuthServer(grpcServer, as.grpcService)`  
-`pb.RegisterAuthServer` is auto generated by grpc `/auth/service/auth_grpc.pb.go`  
-we pass it the `grpcServer` and the `as.grpcService`

`var runErr error`  
-create a run error  
`var wg sync.WaitGroup`  
-create a wait group  
`wg.Add(1)`  
-increment the wait group  
`go func() {}()`  
-spawn a new goroutine  
`defer wg.Done()`  
-finish the waitgroup when this goroutine finishes  
-`runErr = grpcServer.Serve(lis)`  
-start the `grpcServer` and pass it the `lis(tener)`  
-get the run error from the `grpcServer`

`as.config.Log.Printf("auth service: listening: %s", listen)`  
-print to the logger

`<-ctx.Done()`  
-if we recieve a `Done` from the context  
`grpcServer.GracefulStop()`  
-then graceful stop the grpcserver

`wg.Wait()`  
-wait for the server goroutine to finish

`return runErr`  
-return the auth server error

### `type grpcAuthService struct {}`

-internal grpcAuthService struct that implements the gRPC server interface  
-`pb.UnimplementedAuthServer` (generated automatically by grpc)  
-`pool` `*pgxpool.Pool`  
-pool is a reference to the database that we can use for queries

### `func newGrpcService() *grpcAuthService {}`

-constructor  
-returns a grpcAuthService `&grpcAuthService{}`

### `type userRow struct {}`

-`id` `string`  
-`password` `string`  
-`status` `string`

- the type definition for a user in the database

### `func (as *grpcAuthService) Verify(ctx context.Context, in *pb.VerifyRequest) (*pb.VerifyResponse, error) {}`

-arguments: -`ctx context.Context` a context  
-`in *pb.VerifyRequest` a pointer to a protocol buffer verify request

-returns: -`*pb.VerifyResponse` a protocol buffer verify response  
-`error` an error

-verify checks an `input` for authentication validity  
-`log.Printf("verify: id %v, start\n", in.Id)`  
-logs out the verify ?? should we leave this here ??

-`var row userRow`  
-create a row

`err := as.pool.QueryRow(ctx,"SELECT id, password, status FROM public.user WHERE id = $1", in.Id,).Scan(&row.id, &row.password, &row.status)`  
-queries a row from the database id|password|status from `public.user` where `id`  
-id is coming from `in.Id` which is the `*pb.VerifyRequest` (protocol buffer)

`if err != nil {}`  
-if there is an error then:  
-if no rows then the user doesn't exist  
-or if a real error
-then:  
`return &pb.VerifyResponse{State: pb.State_DENY}, nil`  
-return a `pb.State_DENY`

`err = bcrypt.CompareHashAndPassword([]byte(row.password), []byte(in.Password))`  
-"bcrypt require us to compare the input to the hash directly"  
-we compare the database user `password` and the grpc `in.Password`  
-if there is an error  
-mismatch between the database user password and the grpc in.Password
-then log ?? this is a problem why does it say it "is OK" ??
-regardless:  
-`return &pb.VerifyResponse{State: pb.State_DENY}, nil`  
-return a `pb.State_DENY`

-if we reach here it was successful  
-`return &pb.VerifyResponse{State: pb.State_ALLOW}, nil`  
-so we return a `pb.State_ALLOW`
