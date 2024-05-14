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

`Run`  
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
