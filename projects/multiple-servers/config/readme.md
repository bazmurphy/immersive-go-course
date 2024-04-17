# nginx - proof of concept

1. stop the existing `nginx` service

```sh
sudo systemctl stop nginx
```

2. run `nginx` with the custom config

```sh
sudo nginx -c /media/baz/external/coding/immersive-go-course/projects/multiple-servers/config/nginx.conf
```

```sh
baz@baz-pc:~$ sudo nginx -c /media/baz/external/coding/immersive-go-course/projects/multiple-servers/config/nginx.conf
2024/04/17 06:28:23 [notice] 48834#48834: using the "epoll" event method
2024/04/17 06:28:23 [notice] 48834#48834: nginx/1.24.0 (Ubuntu)
2024/04/17 06:28:23 [notice] 48834#48834: OS: Linux 6.5.0-27-generic
2024/04/17 06:28:23 [notice] 48834#48834: getrlimit(RLIMIT_NOFILE): 1024:1048576
2024/04/17 06:28:23 [notice] 48834#48834: start worker processes
...
```

3. run the `api server` (with air)

```sh
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/multiple-servers/cmd/api-server$ air

  __    _   ___
 / /\  | | | |_)
/_/--\ |_| |_| \_ v1.51.0, built with Go go1.22.1

watching .
!exclude tmp
building...
running...
api server running on on port 8081
```

4. run the `static server` (with air)

```sh
baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/multiple-servers/cmd/static-server$ air

  __    _   ___
 / /\  | | | |_)
/_/--\ |_| |_| \_ v1.51.0, built with Go go1.22.1

mkdir /media/baz/external/coding/immersive-go-course/projects/multiple-servers/cmd/static-server/tmp
watching .
!exclude tmp
building...
running...
static server running on port 8082
```

5. test the reverse proxying to the `static server`

```sh
baz@baz-pc:~$ curl http://localhost:8080
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Image gallery</title>
    <link rel="stylesheet" href="style.css" />
    <script src="script.js" defer></script>
  </head>
  <body>
    <div class="wrapper">
      <div class="content" role="main">
        <h1 class="title">Gallery</h1>
        <h2>Sunsets and animals like you've never seen them before.</h2>
        <div class="gallery">Loading images&hellip;</div>
      </div>
    </div>
  </body>
</html>
baz@baz-pc:~$
```

6. test the reverse proxying to the `api server`

```sh
baz@baz-pc:~$ curl http://localhost:8080/api/images.json
[{"title":"Sunset","url":"https://images.unsplash.com/photo-1506815444479-bfdb1e96c566?ixlib=rb-1.2.1\\u0026ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8\\u0026auto=format\\u0026fit=crop\\u0026w=1000\\u0026q=80","alt_text":"Clouds at sunset"},{"title":"Mountain","url":"https://images.unsplash.com/photo-1540979388789-6cee28a1cdc9?ixlib=rb-1.2.1\\u0026ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8\\u0026auto=format\\u0026fit=crop\\u0026w=1000\\u0026q=80","alt_text":"A mountain at sunset"},{"title":"Cat","url":"https://images.unsplash.com/photo-1533738363-b7f9aef128ce?ixlib=rb-1.2.1\u0026ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8\u0026auto=format\u0026fit=crop\u0026w=1000\u0026q=80","alt_text":"A cool cat"}]baz@baz-pc:~$
```
