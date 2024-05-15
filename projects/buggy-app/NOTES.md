# Notes to Understand the App

---

I will start here as it is the entry point of the API Service

## `/cmd/api/main_cmd.go`

-`port` the API Service will listen on (from )

-`passwd, err`  
-read the postgres password from the env or a file

`ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)`  
-setup a context that can be sent a signal to graceful shutdown

`as := auth.New()`
-make a new api service
-passing ita an `auth.Config`  
-`Port:        *port,` port for the API  
-`DatabaseUrl: fmt.Sprintf("postgres://postgres:%s@postgres:5432/app", passwd),` Database URL -`Log:         log.Default(),` A Logger

`if err := as.Run(ctx);`  
-run the API Service passing it the context

`err != nil { log.Fatal(err) }`  
-if there is an error running the API Service then fatal

---

And now move to where `Run()` from above comes from...

## `/api/api.go`

### `DbClient` `interface`

-with 3 methods  
-`QueryRow(context.Context, string, ...interface{}) pgx.Row` used to get an individual note  
-`Query(context.Context, string, ...interface{}) (pgx.Rows, error)` used to get many notes  
-`Close`
-this is the `pool` on the `Service` struct

### `Config` `struct`

-this is what was used above in `/cmd/api/main_cmd.go` -`Port`  
-`Log`  
-`AuthServiceUrl`  
-`DatabaseUrl`
-this is the `config` on the `Service` struct

### `Service` `struct`

-`config` `Config`  
-`authClient` `auth.Client` an interface (with `Close()` and `Verify()`) coming from `/auth/client.go`  
-`pool` `DbClient`

### `New` Service constructor

returns a new `Service`  
and takes in  
-`config`  
-where is the `authClient` `auth.Client` provided ?? (answer it is directly set in `Run()`)  
-where is the `pool` `DbClient` provided ?? (answer it is directly set in `Run()`)

### `Run()`

`func (as *Service) Run(ctx context.Context) error {}`

`listen` gets the port form the `config.Port`

pgsql `pool` create and passed:  
`context`
`as.config.AuthServiceUrl` (API Service config)

`as.pool` is SET at this point from the pool (created just above)

`auth.NewClient(ctx, as.config.AuthServiceUrl)`

`as.authClient` is SET at this point from the client (created just above)

`mux` is created by `as.Handler()`  
-mux is an abbreviation of multiplex  
-which is the root handler that works out where to route the various requests

`server` is created from a new `http.Server{}` taking  
`Addr: listen` (which is the `listen` port created above)  
`Handler: mux` (which is the `mux` created above)

`runErr` created

`wg` wait group created  
-increment the wait group
-spawn a new goroutine (but why do we do this in a new goroutine ?? is it blocking ??)  
`defer wg.Done()` -`server.ListenAndServe()` runs the server

`as.config.Log.Printf` we write a message to the logger

`<-ctx.Done()` if we receive the context Done (where is this Done coming from ?? `pool` and auth `client` both have a context ??)

`server.Shutdown(context.TODO())` (is it ok to use .TODO() here ??) ?? (BUG)

`wg.Wait()` wait for the waitgroup to finish (the single goroutine that is running the server)  
`return runErr` if any

### `Handler()`

`func (as *Service) Handler() http.Handler {}`

`mux := new(http.ServeMux)`  
-this is what determines how to handle the routes  
"`ServeMux` is an HTTP request multiplexer. It matches the URL of each incoming request against a list of registered patterns and calls the handler for the pattern that most closely matches the URL."

`mux.HandleFunc("/1/my/note/", as.wrapAuth(as.authClient, as.handleMyNoteById))`  
-/1/my/note/ uses `handleMyNotById` handler

`mux.HandleFunc("/1/my/notes.json", as.wrapAuth(as.authClient, as.handleMyNotes))`
-/1/my/notes.json uses `handleMyNotes` handler

-both are wrapped in `as.wrapAuth` (which is like a middleware)

`return httplogger.HTTPLogger(mux)` why are we wrapping the `mux` in this logger ?? I don't understand ?? So that it logs all requests presumably ??

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
-why this time do we not use `owner` ?? (like the above) (BUG) ??  
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

---

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

---

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

### `GetNotesForOwner()`

`func GetNotesForOwner(ctx context.Context, conn dbConn, owner string) (Notes, error) {}`

-arguments:  
-`ctx` `context.Context`  
-`conn` `dbConn`  
-`owner` `string`

-returns:  
-`Notes, error`

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

### `GetNoteById()`

`func GetNoteById(ctx context.Context, conn dbConn, id string) (Note, error) {}`

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

### `extractTags()`

`func extractTags(input string) []string {}`

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

### `New()` (Service)

`func New(config Config) *Service {}`

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

### `Verify()` (grpcAuthService)

`func (as *grpcAuthService) Verify(ctx context.Context, in *pb.VerifyRequest) (*pb.VerifyResponse, error) {}`

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

## `/auth/client.go`

### `type Client interface`

-interface for the auth client (with two methods)
`Close() error`  
`Verify(ctx context.Context, id, passwd string) (*VerifyResult, error)`

### `type VerifyResult struct`

-`State` `string`  
-struct for the verify result

### `var ()`

-`StateDeny = pb.State_name[int32(pb.State_DENY)]`  
-`StateAllow = pb.State_name[int32(pb.State_ALLOW)]`  
-the state constants for deny and allow (both strings)
-DENY is 0, ALLOW is 1

### `type GrpcClient struct`

-This is the struct for the gRPC client
-grpcClient is meant to be used by other services to talk with the Auth service  
-`conn` `*grpc.ClientConn` this the grpc client connection  
-`cancel` `context.CancelFunc` this is the context canceller  
-`aC` `pb.AuthClient` protocol buffer auth client  
-`cache` `*cache.Cache[VerifyResult]` we create a cache for the grpc client to use

### `NewClient()`

`func NewClient(ctx context.Context, target string) (*GrpcClient, error)`

-Create a new Client for the auth service

-arguments:  
-`ctx context.Context` the context  
-`target string` target is for `grpc.DialContext`
-Call `Close()` to release resources associated with this Client
-returns:  
`return newClientWithOpts(ctx, target, defaultOpts()...)`

### `Close()` (GrpcClient)

`func (c *GrpcClient) Close() error`

-a method on the GrpcClient (that satisfies the Client interface)  
-`c.cancel()`  
-cancel the context in case the connection is still being formed  
-`return c.conn.Close()`  
-according to grpc.DialContext docs, we still need to call `conn.Close()`  
-Call `Close()` to release resources associated with this Client

### `Verify()` (GrpcClient)

`func (c *GrpcClient) Verify(ctx context.Context, id, passwd string) (*VerifyResult, error)`

-a method on the GrpcClient (that satisfies the Client interface)
-arguments:  
-`ctx context.Context` the context  
-`id string` used to first check the cache and if not then the auth service  
-`passwd string` used to first check the cache and if not then the auth service

-`cacheKey := c.cache.Key(fmt.Sprintf("%s:%s", id, passwd))`  
-check the cache to see if we have this id/passwd combo already there  
-`if v, ok := c.cache.Get(cacheKey); ok { return v, nil }`  
-if we do, return it so we don't contact the auth service twice  
-`res, err := c.aC.Verify(ctx, &pb.VerifyRequest{ Id: id, Password: passwd, })`  
-call the auth service to check the id/password we've been given  
-`if err != nil { return nil, fmt.Errorf("failed to verify: %w", err) }`  
-if there's an error, return it  
-`vR := &VerifyResult{ State: pb.State_name[int32(res.State)], }`  
-looking good: turn this gRPC result into our output type  
-`c.cache.Put(cacheKey, vR)`  
-remember this verify result for next time  
-`return vR, nil`
-return the verify result

### `defaultOpts()` (grpc Dial Options)

`func defaultOpts() []grpc.DialOption`

-returns:  
-`[]grpc.DialOption` from grpc "DialOption configures how we set up the connection"  
-`return []grpc.DialOption{ grpc.WithTransportCredentials(insecure.NewCredentials()), }`
-"WithTransportCredentials returns a DialOption which configures a connection level security credentials (e.g., TLS/SSL). This should not be used together with WithCredentialsBundle."  
-TODO: insecure connection should move to TLS

### `func newClientWithOpts(ctx context.Context, target string, opts ...grpc.DialOption) (*GrpcClient, error)`

-Use this function in tests to configure the underlying client with options

-arguments:  
-`ctx context.Context` the context  
-`target string` target is for `grpc.DialContext`  
-`opts ...grpc.DialOption` spread in a slice of `grpc.Dialoptions`

-returns:  
-`*GrpcClient` our own GrpcClient that other services uses to talk with the Auth service  
-`error`

-`ctx, cancel := context.WithCancel(ctx)`  
-Wrapping the context WithCancel allows us to cancel the connection if the caller chooses to immediately `Close()` the Client  
-`conn, err := grpc.DialContext(ctx, target, opts...)`  
-Create the gRPC connection -`if err != nil { return nil, fmt.Errorf("failed to create client: %w", err) }`  
-If there's an error, return it  
-`return &GrpcClient{ conn: conn, cancel: cancel, aC: pb.NewAuthClient(conn), cache: cache.New[VerifyResult](), }, nil`  
-`conn` is `*grpc.ClientConn`  
-`cancel` is `context.CancelFunc`  
-`aC` is `pb.AuthClient` protocol buffer AuthClient  
-`cache` is `*cache.Cache[VerifyResult]` a new cache with a key as `VerifyResult`
-Return the new gRPC client

### `type MockClient struct`

-This is the struct for the mock client  
-`result` `*VerifyResult`

### `NewMockClient` (for tests)

`func NewMockClient(result *VerifyResult) *MockClient`

-arguments:  
-`result *VerifyResult`  
-`return &MockClient{ result: result }`  
-create a new mock client with the given verify result

### `Close` (on MockClient for tests)

`func (ac *MockClient) Close() error`

-`return nil`  
-this is a "no-op" for the mock client

### `Verify` (on MockClient for tests)

`func (ac *MockClient) Verify(ctx context.Context, id, passwd string) (*VerifyResult, error)`

-arguments:  
-`ctx.Context` the context  
-`id` the id for auth  
-`passwd` the password for auth

-returns:  
`*VerifyResult`  
`error`

-`return ac.result, nil`  
-Return the mock verify result  
-Use this in tests to Mock out the client

## `/auth/cache/cache.go`

### `type Key [16]byte`

-defines the `Key` type as an array of 16 bytes

### `type Entry[Value any] struct`

-`value` `*Value`  
-defines a generic `Entry` struct that holds a pointer to a value of the specified type `Value`

### `type Cache[Value any] struct`

-`entries` `*sync.Map`  
-defines a generic `Cache` struct that holds a pointer to a `sync.Map` containing entries of the specified type `Value`

### New() (create a new Cache)

`func New[Value any]() *Cache[Value]`

-returns:  
-`*Cache[Value]`

-`return &Cache[Value]{ entries: &sync.Map{}, }`  
-constructor function that creates and returns a new `Cache` instance with an empty `sync.Map` for entries

### Key() (Cache) (convert the key to a hashed key)

`func (c *Cache[V]) Key(k string) Key`

-arguments:  
-`k string` the key to hash

-returns:  
-`Key   type Key [16]byte`

-`return md5.Sum([]byte(k))`
-method takes a string `k` and returns its MD5 hash as a `Key`

### Get() (Cache) (lookup the key, return the entry and found boolean)

`func (c *Cache[Value]) Get(k Key) (*Value, bool)`

-arguments:  
-`k string` the key to get

-returns:  
-`*Value` the value of the key  
-`bool` if the entry was found

-`if value, ok := c.entries.Load(k); ok`  
-attempt to load the entry with the given `Key` from the `entries` map -`Load` is from `sync.Map` "Load returns the value stored in the map for a key, or nil if no value is present. The ok result indicates whether value was found in the map."

-`if entry, ok := value.(Entry[Value]); ok`  
-if the entry exists and can be type-asserted to `Entry[Value]`, return the value and `true`

-`return nil, false`  
-if the entry doesn't exist or cannot be type-asserted, return `nil` and `false`

### Put() (Cache) (create or update a cache key)

`func (c *Cache[Value]) Put(k Key, v *Value)`

-arguments:  
-`k Key`  
-`v *Value`

-`c.entries.Store(k, Entry[Value]{ value: v, })`  
-`Store` is from `sync.Map` "Store sets the value for a key"
-create a new `Entry[Value]` with the given value `v` and store it in the `entries` map with the given `Key` `k`.

## `auth/service/auth.proto` (VerifyRequest (id and password) -> VerifyResponse (State Deny/Allow))

`package service`
-the package name for the generated grpc Go code

`service Auth {}`
-the `Auth` service definition  
-`rpc Verify(VerifyRequest) returns (VerifyResponse) {}`
-an RPC method `Verify` that takes a `VerifyRequest` message as input and returns a `VerifyResponse` message
-"Callers should deny access to resources unless the Result is ALLOW" ?? How can i verify this is happening ??

`message VerifyRequest { string id = 1;  string password = 2; }`  
-a message definition  
-`string id = 1` a string field `id` with field number `1`  
-`string password = 2` a string field `password` with field number `2`

`message VerifyResponse { State state = 1; }`  
-a message definition  
-`State state = 1` a field `state` of type `State` with field number `1`

-`enum State { DENY = 0; ALLOW = 1; }`  
-a enumeration named `State`  
-`DENY` with value `0`  
-`ALLOW` with value `1`

---

## `/util/authuserctx/authctx.go`

-this package basically handles adding and retrieving a user id to/from a context

`type key int`
-custom key type as an integer

`const authenticatedIdKey key = 0`
-what is a "user identifier" in this case ??  
-"this constant serves as the key for storing and retrieving the authenticated user's ID in the context"

-this is done so we can lookup key 0 which relates to the authenticatedIdKey  
-if for example we had other context values we need some way to make sure they are all separate, so this is why we give it a fixed key  
-if we wanted to add another value, we could use const someOtherKey = 1

## `NewAuthenticatedContext()`

`func NewAuthenticatedContext(ctx context.Context, id string) context.Context {}`

-arguments:  
`ctx context.Context` the context  
`id string` the user id

-returns:  
`context.Context` a context

`return context.WithValue(ctx, authenticatedIdKey, id)`
-returns a context (that has a value that can be accessed by other parts of the app)
-the key is `authenticatedIdKey` (the constant from above)  
-the value is `id` `string`

-this function is for creating a new context with the user id added to it

## `FromAuthenticatedContext()`

`func FromAuthenticatedContext(ctx context.Context) (string, bool) {}`

-arguments: -`ctx context.Context`

-returns:
`string` the user id  
`bool` successful or not in getting the user id

`id, ok := ctx.Value(authenticatedIdKey).(string)`  
-get the value of the `authenticatedIdKey`  
-type assert the value to a `string`

`return id, ok`  
-return the user id (string) and ok (success) (boolean)  
-user user id string and true if found  
-empty string and false if not found

-this function is for getting the user id from a context

---

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

app=# SELECT * from "user";
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
