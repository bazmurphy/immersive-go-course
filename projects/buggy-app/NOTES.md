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

### `New()` (create a new Cache)

`func New[Value any]() *Cache[Value]`

-returns:  
-`*Cache[Value]`

-`return &Cache[Value]{ entries: &sync.Map{}, }`  
-constructor function that creates and returns a new `Cache` instance with an empty `sync.Map` for entries

### `Key()` (Cache) (convert the key to a hashed key)

`func (c *Cache[V]) Key(k string) Key`

-arguments:  
-`k string` the key to hash

-returns:  
-`Key   type Key [16]byte`

-`return md5.Sum([]byte(k))`
-method takes a string `k` and returns its MD5 hash as a `Key`

### `Get()` (Cache) (lookup the key, return the entry and found boolean)

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

### `Put()` (Cache) (create or update a cache key)

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

## `/util/basic_auth.go`

### `BasicAuthValue()`

`func BasicAuthValue(id, password string) string {}`

- generates a Value for Basic Auth Header

-arguments:  
-`id string`  
-`password string`

-returns:  
`string` base64 encoded id:password combination

### `BasicAuthHeaderValue()`

-generates a Basic Auth Header Value

-arguments:  
-`id string`  
-`password string`

-returns:  
`string` Basic id:password(as base64)

---

## `/util/postgres.go`

### `ReadPasswd()`

`func ReadPasswd() (string, error) {}`

-returns:
`string` the postgresql database password
`error` error if we can't get it

-when we `os.ReadFile` we directly dump the err out, is that ok, shouldn't we add more information ??  
-this `pwdFile` could be multi-line ??

---

## `/util/service.go`

### `MarshalWithIndent()`

`func MarshalWithIndent(data interface{}, indent string) ([]byte, error) {}`

-arguments:  
-`data interface{}` interface{} is `any`  
-`indent string` the amount to indent by (restricted from 1 to 10)

-try to convert the `indent` to an integer  
-`if i, err := strconv.Atoi(indent)`  
-a `err == nil` check is done  
-but the `err` here is never handled ?? (BUG)

-if `ident` is a number between 1 and 10 then it uses `json.MarshalIndent`  
-otherwise it uses `json.Marshal`

`return b, nil`
-return the bytes slice, and nil err

---

I want to look at the container/build stuff now...

## `Dockerfile`

`# This Dockerfile contains all code for the entire repository.`  
`# To run a different executable, supply a different command.`  
`# To avoid the "wait for Postgres" feature, supply a different entrypoint.`

`WORKDIR /app`  
working directory on the container is `/app`

copy across the `go.mod` and `go.sum` files

run a `go mod download`

copy across the various folders (into `/app`)  
`api`  
`auth`  
`cmd`  
`migrations`  
`util`

make the `/out` directory

build the binaries and put them in `/out`  
i assume `./...` is recursively? (like `go test`)

copy `bin` to `/bin`  
(these contain the scripts `docker-entrypoint.sh` and `wait-for-it.sh`)

copy `migrations` to `/migrations`

expose the container port `80`

"The entrypoint will, by default, wait for postgres to become available at `postgres://postgres:5432` before running the command that follows"

`ENTRYPOINT [ "/bin/docker-entrypoint.sh" ]`

this entrypoint defines what runs when the container starts

---

## `/bin/docker-entrypoint.sh`

`set -e`
"this line enables the "exit immediately" option for the shell. If any command in the script fails (returns a non-zero exit code), the script will immediately abort execution. This helps in catching and propagating errors."

`/bin/wait-for-it.sh postgres:5432 -t 60 --`

"this line executes the `wait-for-it.sh` script, which is a common utility script used to wait for a specific host and port to become available before proceeding. In this case, it waits for the `postgres` host on port `5432` to be accessible. The `-t 60` option specifies a timeout of `60` seconds, so if the `postgres` host is not available within that time, the script will abort. The `--` at the end is used to separate the options of `wait-for-it.sh` from the rest of the command."

`exec "$@"`

"this line replaces the current shell process with the command specified by the arguments passed to the script. "$@" expands to all the arguments passed to the script. The exec command ensures that the specified command becomes the main process of the container, replacing the script itself."

so the commands passed to the dockerfile (or docker compose, or the makefile running all of that)

for example: `docker run someimage command1 command2 command2`

the commands will be run via `"$@"` at the end of the script

---

## `docker-compose.yml`

5 services: `postgres` `migrate` `auth` `api` `test`

### `postgres`

`image: postgres` get the official postgres image
`restart: always` always restart
`volumes:` mounts volumes to the container
`/tmp/buggy-app-data` is mounted as the data storage volume
`/volumes/secrets` is mounted as a read-only volume for the secrets
`/volumes/init` is mounted as a read-only volume for the initialisation scripts
`environment:` sets environment variables for the container
`POSTGRES_PASSWORD_FILE=/run/secrets/postgres-passwd` this is the file containing the postgres password
`POSTGRES_HOST=postgres` this sets the host name for the postgres service
`ports` maps the container port 5432 to the host port 5432

### `migrate`

`build: .` builds the docker image using the Dockerfile
`depends_on : postgres` it depends on the `postgres` service above
`volumes:` mounts volumes to the container
`/volumes/secrets` is mounted as a read-only volume for the secrets
`environment:` sets environment variables for the container
`POSTGRES_PASSWORD_FILE=/run/secrets/postgres-passwd` this is the file containing the postgres password
`command: /out/migrate --path /migrations up` this is the command that runs when the container starts (it runs the database migrations)
`profiles: ["migrate"]` this service is linked to the `migrate` `profile`

### `auth`

`build: .` builds the docker image using the Dockerfile
`ports: 127.0.0.1:8080:80` maps the containers port `80` to the host's port 8080 (why do we need 127.0.0.1 ??)
`depends_on: postgres` it depends on the `postgres` service above
`volumes:` mounts volumes to the container
`/volumes/secrets` is mounted as a read-only volume for the secrets
`environment:` sets environment variables for the container
`POSTGRES_PASSWORD_FILE=/run/secrets/postgres-passwd` this is the file containing the postgres password
`command: /out/auth` this is the command that runs when the container starts (it runs the auth service)

### `api`

`build: .` builds the docker image using the Dockerfile
`ports: 127.0.0.1:8090:80` maps the containers port `80` to the host's port 8080 (why do we need 127.0.0.1 ??)
`depends_on:` it depends on the `postgres` and `auth` services above
`- postgres`
`- auth`
`volumes:` mounts volumes to the container
`/volumes/secrets` is mounted as a read-only volume for the secrets
`environment:` sets environment variables for the container
`POSTGRES_PASSWORD_FILE=/run/secrets/postgres-passwd` this is the file containing the postgres password
`command: /out/auth` this is the command that runs when the container starts (it runs the api service)

### `test`

`build: .` builds the docker image using the Dockerfile
`depends_on: postgres` it depends on the `postgres` service above
`volumes:` mounts volumes to the container
`/volumes/secrets` is mounted as a read-only volume for the secrets
`command: go test /app/...` his is the command that runs when the container starts (it runs the tests)
`environment:` sets environment variables for the container (why is this after command ??)
`POSTGRES_PASSWORD_FILE=/run/secrets/postgres-passwd` this is the file containing the postgres password

---

## `Makefile`

`.PHONY: protoc volumes volumes-reset test run`

- this declares the targets that are not associated with file names. It helps avoid conflicts with files that might have the same name as the targets.

`volumes/secrets/postgres-passwd:`

- this target creates a random password for the Postgres database and stores it in the `volumes/secrets/postgres-passwd` file.
- it first creates the `volumes/secrets` directory using `mkdir -p`.
- then, it generates a random password using `openssl rand -hex 24`, removes any newline characters with `tr -d '\\n'`, and saves it to the `volumes/secrets/postgres-passwd` file.

`volumes: volumes/secrets/postgres-passwd`

- this target depends on the `volumes/secrets/postgres-passwd` target and creates the `/tmp/buggy-app-data` directory using `mkdir -p`.
- the `/tmp/buggy-app-data` directory is used for storing the Postgres database data.

`volumes-reset:`

- this target is used to completely reset the database state by removing the `/tmp/buggy-app-data` directory using `rm -rf`.
- it should be used with caution and not run while the containers are running.

`protoc: auth/service/auth.proto`

- this target compiles the `auth/service/auth.proto` protobuf file using the `protoc` command.
- it generates Go code and gRPC code based on the protobuf definitions.

`test: volumes`

- this target runs the tests for the application.
- it first builds the Go code using `go build ./...` to ensure that the code compiles.
- it then builds the Docker containers for the test profile using `docker compose --profile test build`.
- finally, it runs the tests using `docker compose run -T test`. The `-T` flag forces Docker not to allocate a TTY, which is important for pre-commit hooks.

`build: volumes`

- this target builds the application.
- it first builds the Go code using `go build ./...` to ensure that the code compiles.
- it then builds the Docker containers for the run profile using `docker compose --profile run build`.

`migrate: volumes`

- this target runs the database migrations.
- it builds the Docker containers for the migrate profile using `docker compose --profile migrate build`.
- it then runs the migrations using `docker compose run migrate`.

`run:`

- this target runs the application using `docker compose --profile run up`.

`run-database:`

- this target starts only the Postgres database container using `docker compose up`. (BUG ?? this is starting all the containers not just the database)

`build-run: | build run`

- this target first runs the `build` target and then the `run` target.
- the `|` symbol is used to specify that `build` and `run` are order-only prerequisites, meaning they will be executed in the specified order.

`migrate-local:`

- this target runs the database migrations locally without using Docker.
- it sets the `POSTGRES_PASSWORD_FILE` environment variable to `volumes/secrets/postgres-passwd`.
- it then runs the `migrate` command using `go run ./cmd/migrate --hostport localhost:5432 --path migrations up`.

`migrate-local-down:`

- this target rolls back the database migrations locally without using Docker.
- it sets the`POSTGRES_PASSWORD_FILE`environment variable to`volumes/secrets/postgres-passwd`.
- it then runs the `migrate`command using`go run ./cmd/migrate --hostport localhost:5432 --path migrations down`.

---
