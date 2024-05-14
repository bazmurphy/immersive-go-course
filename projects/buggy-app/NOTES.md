# Notes to Understand the App

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

## `/api/api.go`

`DbClient` `interface` with 3 methods  
-QueryRow  
-Query  
-Close

`Config` `struct` (this is what was used above in `/cmd/api/main_cmd.go`)  
-Port  
-Log  
-AuthServiceUrl  
-DatabaseUrl

`Service` `struct`  
-config (Config)  
-authClient (auth.Client)  
-pool (DbClient)

`New` Service constructor  
returns a new Service  
and takes in  
-config  
-but where is the authClient (auth.Client) ??  
-where is the pool (DbClient) ??

### `Run()`

-listen gets the port form the `config.Port`

-pgsql `pool` created, passed:  
-context  
-`as.config.AuthServiceUrl` (API Service config)

-`as.pool` is SET at this point from the pool (created just above)

`auth.NewClient(ctx, as.config.AuthServiceUrl)`

-`as.authClient` is SET at this point from the client (created just above)

-`mux` is created by `as.Handler()`  
mux is an abbreviation of multiplex  
which is the root handler that works out where to route the various requests

-`server` is created from a new `http.Server{}` taking  
-`Addr: listen` (which is the `listen` port created above)  
-`Handler: mux` (which is the `mux` created above)

-`runErr` created

-`wg` wait group created  
-add 1 to the wg
-spawn a new goroutine (but why do we do this in a new goroutine ??)  
-`defer wg.Done()` -`server.ListenAndServe()` runs the server

-`as.config.Log.Printf` we write a message to the logger

-`<-ctx.Done()` if we receive the context Done

-`server.Shutdown(context.TODO())` (is it ok to use .TODO() here ??)

-`wg.Wait()` wait for the waitgroups to finish (the single goroutine that is running the server)  
-`return runErr` if any

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

-`ctx := r.Context()`  
-"Context returns the request's context."

-`owner, ok := authuserctx.FromAuthenticatedContext(ctx)`  
-get the authenticated user from the context -- this will have been written earlier  
-if not ok return an http error

-`notes, err := model.GetNotesForOwner(ctx, as.pool, owner)`  
-use the "model" layer to get a list of the owner's notes  
-if not ok return http error

-create a `response` struct (with the `model.Notes`)

-`res, err := util.MarshalWithIndent(response, "")`  
-why is there an empty string here ?? shouldn't it be the indent amount ??  
-if not ok print error

-`w.Header().Add("Content-Type", "text/json")`
-write the header
-this should be `application/json` !! [BUG]

-`w.Write(res)`  
-write the body

Q: why are some of these logging to the logger and some just to Printf ??

### `handleMyNoteById()`

`func (as *Service) handleMyNoteById(w http.ResponseWriter, r *http.Request) {}`

-`ctx := r.Context()`  
-"Context returns the request's context."

-`_, ok := authuserctx.FromAuthenticatedContext(ctx)`  
-why this time do we not use `owner` ?? [BUG] (like the above)  
-if not ok return http error

-`id := strings.Replace(path.Base(r.URL.Path), ".json", "", 1)`  
-get the id from the url path  
-is this process of stripping out the id correct ??  
-if no id then error

-`note, err := model.GetNoteById(ctx, as.pool, id)`  
-use the "model" layer to get a list of the owner's notes  
-how do we get the owner here?
-and why are we allowing to get all the owner's notes? can everyone access this ?? look at the auth after this ??
-if err then error (failed to get the note)

-`response := struct`  
-build a response with `model.Note`

-`res, err := util.MarshalWithIndent(response, "")`  
-again we are passing an empty string into the `util.MarshalWithIdent`... why ?? [BUG] ??

-`w.Header().Add("Content-Type", "text/json")`
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
-what is `r.Context()` ?

-`id, passwd, ok := r.BasicAuth()`
-if not ok then return http error

-`result, err := client.Verify(ctx, id, passwd)`  
-use the auth client to check if this id/password combo is approved  
-`client.Verify()` is coming from `/auth/client.go`  
-if err then return http error

-`if result.State != auth.StateAllow`  
-unless we get an allow then deny  
-`.State` is coming from `/auth/client.go`  
-`auth.StateAllow` is coming from `/auth/client.go`

-`ctx = authuserctx.NewAuthenticatedContext(ctx, id)`
-add the ID to the context and call the inner handler  
-`NewAuthenticatedContext` is coming from `/util/authuserctx`

-`handler(w, r.WithContext(ctx))`
-returns a reader with `ctx` from the line above^
-returns a `handler` - i assume this is an implicit return?
