# nginx - Install, Setup, Configure

## Install

Install nginx

```sh
sudo apt install nginx
```

Check the version

```sh
baz@baz-pc:~$ nginx -v
nginx version: nginx/1.24.0 (Ubuntu)
```

Check the status

```sh
baz@baz-pc:~$ sudo systemctl status nginx
● nginx.service - A high performance web server and a reverse proxy server
     Loaded: loaded (/lib/systemd/system/nginx.service; enabled; preset: enabled)
     Active: active (running) since Mon 2024-04-08 13:18:18 BST; 8min ago
       Docs: man:nginx(8)
   Main PID: 17772 (nginx)
      Tasks: 17 (limit: 18982)
     Memory: 11.6M
        CPU: 64ms
     CGroup: /system.slice/nginx.service
             ├─17772 "nginx: master process /usr/sbin/nginx -g daemon on; master_process on;"
             ├─17774 "nginx: worker process"
             ├─17775 "nginx: worker process"
             ├─17776 "nginx: worker process"
             ├─17777 "nginx: worker process"
             ├─17778 "nginx: worker process"
             ├─17779 "nginx: worker process"
             ├─17780 "nginx: worker process"
             ├─17781 "nginx: worker process"
             ├─17782 "nginx: worker process"
             ├─17783 "nginx: worker process"
             ├─17784 "nginx: worker process"
             ├─17785 "nginx: worker process"
             ├─17786 "nginx: worker process"
             ├─17787 "nginx: worker process"
             ├─17788 "nginx: worker process"
             └─17789 "nginx: worker process"

Apr 08 13:18:18 baz-pc systemd[1]: Starting nginx.service - A high performance web server and a reverse proxy server...
Apr 08 13:18:18 baz-pc systemd[1]: Started nginx.service - A high performance web server and a reverse proxy server.
baz@baz-pc:~$
```

Visit http://127.0.0.1 in Web Browser

I see an Apache 2 Default Page from Ubuntu... Why?

"If you see the Apache2 Default Page with the Ubuntu logo when accessing `http://127.0.0.1`, it means that Apache2 web server is already installed and running on your Ubuntu machine. To switch from Apache2 to Nginx, you'll need to stop the Apache2 service and then proceed with configuring Nginx."

Stop the Apache2 service:

```sh
sudo systemctl stop apache2
```

Disable the Apache2 service to prevent it from starting automatically on system boot:

```sh
sudo systemctl disable apache2
```

```sh
baz@baz-pc:~$ sudo systemctl stop apache2
baz@baz-pc:~$ sudo systemctl disable apache2
Synchronizing state of apache2.service with SysV service script with /lib/systemd/systemd-sysv-install.
Executing: /lib/systemd/systemd-sysv-install disable apache2
Removed "/etc/systemd/system/multi-user.target.wants/apache2.service".
baz@baz-pc:~$
```

Check the Apache2 service status

```sh
sudo systemctl status apache2
```

```sh
× apache2.service - The Apache HTTP Server
     Loaded: loaded (/lib/systemd/system/apache2.service; disabled; preset: enabled)
     Active: failed (Result: exit-code) since Mon 2024-04-08 13:18:24 BST; 12min ago
       Docs: https://httpd.apache.org/docs/2.4/
        CPU: 15ms

Apr 08 13:18:24 baz-pc apachectl[19773]: AH00558: apache2: Could not reliably determine the server's fully qualified domain name, using 127.0.1.1. Set the 'ServerName' directive globally to suppress this message
Apr 08 13:18:24 baz-pc apachectl[19773]: (98)Address already in use: AH00072: make_sock: could not bind to address [::]:80
Apr 08 13:18:24 baz-pc apachectl[19773]: (98)Address already in use: AH00072: make_sock: could not bind to address 0.0.0.0:80
Apr 08 13:18:24 baz-pc apachectl[19773]: no listening sockets available, shutting down
Apr 08 13:18:24 baz-pc apachectl[19773]: AH00015: Unable to open logs
Apr 08 13:18:24 baz-pc apachectl[19770]: Action 'start' failed.
Apr 08 13:18:24 baz-pc apachectl[19770]: The Apache error log may have more information.
Apr 08 13:18:24 baz-pc systemd[1]: apache2.service: Control process exited, code=exited, status=1/FAILURE
Apr 08 13:18:24 baz-pc systemd[1]: apache2.service: Failed with result 'exit-code'.
Apr 08 13:18:24 baz-pc systemd[1]: Failed to start apache2.service - The Apache HTTP Server.
```

Enable the Nginx service to start automatically on system boot:

```sh
sudo systemctl enable nginx
```

```sh
baz@baz-pc:~$ sudo systemctl enable nginx
Synchronizing state of nginx.service with SysV service script with /lib/systemd/systemd-sysv-install.
Executing: /lib/systemd/systemd-sysv-install enable nginx
baz@baz-pc:~$
```

Visit [http://127.0.0.1](http://127.0.0.1) again but can still see the Apache2 Ubuntu Default Page? Why?

"The default page is typically stored in the `/var/www/html` directory on Ubuntu systems."

So have a look at the contents of `/var/www/html`

```sh
baz@baz-pc:~$ ls -la /var/www/html/
total 24
drwxr-xr-x 2 root root  4096 Apr  8 13:18 .
drwxr-xr-x 3 root root  4096 Apr  8 13:18 ..
-rw-r--r-- 1 root root 10671 Apr  8 13:18 index.html
-rw-r--r-- 1 root root   615 Apr  8 13:18 index.nginx-debian.html
```

Check each html file's contents...

```sh
baz@baz-pc:~$ cat /var/www/html/index.html | head
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
  <!--
    Modified from the Debian original for Ubuntu
    Last updated: 2022-03-22
    See: https://launchpad.net/bugs/1966004
  -->
  <head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    <title>Apache2 Ubuntu Default Page: It works</title>
baz@baz-pc:~$
```

This is the "Apache2 Ubuntu Default Page"

```sh
baz@baz-pc:~$ cat /var/www/html/index.nginx-debian.html | head
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
html { color-scheme: light dark; }
body { width: 35em; margin: 0 auto;
font-family: Tahoma, Verdana, Arial, sans-serif; }
</style>
</head>
```

And this is the "Welcome to nginx!" page

So I visit [http://127.0.0.1/index.nginx-debian.html](http://127.0.0.1/index.nginx-debian.html) explicitly and can now see the nginx default page

This explains why I see the "Apache2 Ubuntu Default Page" when accessing [http://127.0.0.1](http://127.0.0.1)

"Even though Apache2 is not running, the `index.html` file in the `/var/www/html` directory contains the Apache2 default page content, and Nginx is serving that file as the default page."

---

## Configure nginx

To configure Nginx, you'll need to edit the configuration files located in the `/etc/nginx/` directory. The main configuration file is `/etc/nginx/nginx.conf`.

Look at the `nginx` directory

```sh
baz@baz-pc:~$ ls -la /etc/nginx/
total 84
drwxr-xr-x   8 root root  4096 Apr  8 13:18 .
drwxr-xr-x 157 root root 12288 Apr  8 13:18 ..
drwxr-xr-x   2 root root  4096 Jun 28  2023 conf.d
-rw-r--r--   1 root root  1125 Jun 27  2023 fastcgi.conf
-rw-r--r--   1 root root  1055 Jun 27  2023 fastcgi_params
-rw-r--r--   1 root root  2837 Jun 27  2023 koi-utf
-rw-r--r--   1 root root  2223 Jun 27  2023 koi-win
-rw-r--r--   1 root root  4338 Jun 27  2023 mime.types
drwxr-xr-x   2 root root  4096 Jun 28  2023 modules-available
drwxr-xr-x   2 root root  4096 Jun 28  2023 modules-enabled
-rw-r--r--   1 root root  1446 Jun 27  2023 nginx.conf
-rw-r--r--   1 root root   180 Jun 27  2023 proxy_params
-rw-r--r--   1 root root   636 Jun 27  2023 scgi_params
drwxr-xr-x   2 root root  4096 Apr  8 13:18 sites-available
drwxr-xr-x   2 root root  4096 Apr  8 13:18 sites-enabled
drwxr-xr-x   2 root root  4096 Apr  8 13:18 snippets
-rw-r--r--   1 root root   664 Jun 27  2023 uwsgi_params
-rw-r--r--   1 root root  3071 Jun 27  2023 win-utf
baz@baz-pc:~$
```

"The main Nginx configuration file is located at `/etc/nginx/nginx.conf`.
Site-specific configurations are stored in the `/etc/nginx/sites-available` directory.
Enabled site configurations are symlinked to the `/etc/nginx/sites-enabled` directory."

Look at the `nginx.conf` file

```sh
baz@baz-pc:~$ cat /etc/nginx/nginx.conf
user www-data;
worker_processes auto;
pid /run/nginx.pid;
error_log /var/log/nginx/error.log;
include /etc/nginx/modules-enabled/*.conf;

events {
	worker_connections 768;
	# multi_accept on;
}

http {

	##
	# Basic Settings
	##

	sendfile on;
	tcp_nopush on;
	types_hash_max_size 2048;
	# server_tokens off;

	# server_names_hash_bucket_size 64;
	# server_name_in_redirect off;

	include /etc/nginx/mime.types;
	default_type application/octet-stream;

	##
	# SSL Settings
	##

	ssl_protocols TLSv1 TLSv1.1 TLSv1.2 TLSv1.3; # Dropping SSLv3, ref: POODLE
	ssl_prefer_server_ciphers on;

	##
	# Logging Settings
	##

	access_log /var/log/nginx/access.log;

	##
	# Gzip Settings
	##

	gzip on;

	# gzip_vary on;
	# gzip_proxied any;
	# gzip_comp_level 6;
	# gzip_buffers 16 8k;
	# gzip_http_version 1.1;
	# gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;

	##
	# Virtual Host Configs
	##

	include /etc/nginx/conf.d/*.conf;
	include /etc/nginx/sites-enabled/*;
}


#mail {
#	# See sample authentication script at:
#	# http://wiki.nginx.org/ImapAuthenticateWithApachePhpScript
#
#	# auth_http localhost/auth.php;
#	# pop3_capabilities "TOP" "USER";
#	# imap_capabilities "IMAP4rev1" "UIDPLUS";
#
#	server {
#		listen     localhost:110;
#		protocol   pop3;
#		proxy      on;
#	}
#
#	server {
#		listen     localhost:143;
#		protocol   imap;
#		proxy      on;
#	}
#}
baz@baz-pc:~$
```

Look at the `/etc/nginx/sites-available` directory

```sh
baz@baz-pc:~$ ls -la /etc/nginx/sites-available/
total 12
drwxr-xr-x 2 root root 4096 Apr  8 13:18 .
drwxr-xr-x 8 root root 4096 Apr  8 13:18 ..
-rw-r--r-- 1 root root 2412 Jun 27  2023 default
```

Look at the `default` file in there

```sh
baz@baz-pc:~$ cat /etc/nginx/sites-available/default
##
# You should look at the following URL's in order to grasp a solid understanding
# of Nginx configuration files in order to fully unleash the power of Nginx.
# https://www.nginx.com/resources/wiki/start/
# https://www.nginx.com/resources/wiki/start/topics/tutorials/config_pitfalls/
# https://wiki.debian.org/Nginx/DirectoryStructure
#
# In most cases, administrators will remove this file from sites-enabled/ and
# leave it as reference inside of sites-available where it will continue to be
# updated by the nginx packaging team.
#
# This file will automatically load configuration files provided by other
# applications, such as Drupal or Wordpress. These applications will be made
# available underneath a path with that package name, such as /drupal8.
#
# Please see /usr/share/doc/nginx-doc/examples/ for more detailed examples.
##

# Default server configuration
#
server {
	listen 80 default_server;
	listen [::]:80 default_server;

	# SSL configuration
	#
	# listen 443 ssl default_server;
	# listen [::]:443 ssl default_server;
	#
	# Note: You should disable gzip for SSL traffic.
	# See: https://bugs.debian.org/773332
	#
	# Read up on ssl_ciphers to ensure a secure configuration.
	# See: https://bugs.debian.org/765782
	#
	# Self signed certs generated by the ssl-cert package
	# Don't use them in a production server!
	#
	# include snippets/snakeoil.conf;

	root /var/www/html;

	# Add index.php to the list if you are using PHP
	index index.html index.htm index.nginx-debian.html;

	server_name _;

	location / {
		# First attempt to serve request as file, then
		# as directory, then fall back to displaying a 404.
		try_files $uri $uri/ =404;
	}

	# pass PHP scripts to FastCGI server
	#
	#location ~ \.php$ {
	#	include snippets/fastcgi-php.conf;
	#
	#	# With php-fpm (or other unix sockets):
	#	fastcgi_pass unix:/run/php/php7.4-fpm.sock;
	#	# With php-cgi (or other tcp sockets):
	#	fastcgi_pass 127.0.0.1:9000;
	#}

	# deny access to .htaccess files, if Apache's document root
	# concurs with nginx's one
	#
	#location ~ /\.ht {
	#	deny all;
	#}
}


# Virtual Host configuration for example.com
#
# You can move that to a different file under sites-available/ and symlink that
# to sites-enabled/ to enable it.
#
#server {
#	listen 80;
#	listen [::]:80;
#
#	server_name example.com;
#
#	root /var/www/example.com;
#	index index.html;
#
#	location / {
#		try_files $uri $uri/ =404;
#	}
#}
baz@baz-pc:~$
```

Look at the `/etc/nginx/sites-enabled` directory

```sh
baz@baz-pc:~$ ls -la /etc/nginx/sites-enabled/
total 8
drwxr-xr-x 2 root root 4096 Apr  8 13:18 .
drwxr-xr-x 8 root root 4096 Apr  8 13:18 ..
lrwxrwxrwx 1 root root   34 Apr  8 13:18 default -> /etc/nginx/sites-available/default
```

"`default -> /etc/nginx/sites-available/default`: This line shows a symbolic link named `default` in the `/etc/nginx/sites-enabled/` directory. The link points to the file `/etc/nginx/sites-available/default`.

In the context of Nginx configuration, the `/etc/nginx/sites-enabled/` directory is used to store the active site configurations. The `default` symbolic link points to the default site configuration file located in the `/etc/nginx/sites-available/` directory. This allows Nginx to easily enable or disable site configurations by creating or removing symbolic links in the `sites-enabled` directory."

## Create a new site configuration

Make a new empty site configuration

```sh
baz@baz-pc:~$ sudo touch /etc/nginx/sites-available/test.com
```

Check the empty configuration was created

```sh
baz@baz-pc:~$ ls -la /etc/nginx/sites-available
total 12
drwxr-xr-x 2 root root 4096 Apr  8 14:06 .
drwxr-xr-x 8 root root 4096 Apr  8 13:18 ..
-rw-r--r-- 1 root root 2412 Jun 27  2023 default
-rw-r--r-- 1 root root    0 Apr  8 14:06 test.com
baz@baz-pc:~$
```

Open the configuration in VSCode and add the following

```sh
baz@baz-pc:~$ code /etc/nginx/sites-available/test.com
baz@baz-pc:~$
```

```nginx
server {
    listen 80;
    server_name test.com;
    root /var/www/test.com/html;
    index index.html;

    location / {
        try_files $uri $uri/ =404;
    }
}
```

Create the "document root directory"

"In the context of the `mkdir` command in Unix-like operating systems (such as Linux), the `-p` flag stands for "parents" or "parent directories." When you use `mkdir` with the `-p` flag followed by a path, it will create the specified directory and also create any necessary parent directories along the path that do not already exist."

```sh
baz@baz-pc:~$ sudo mkdir -p /var/www/test.com/html
baz@baz-pc:~$
```

Create an index.html file in the document root

```sh
baz@baz-pc:~$ sudo touch /var/www/test.com/html/index.html
baz@baz-pc:~$
```

Edit it in VSCode to add some basic html

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Baz test.com nginx</title>
    <style>
      body {
        display: grid;
        place-content: center;
        text-align: center;
      }
      img {
        margin: 0 auto;
      }
    </style>
  </head>
  <body>
    <h1>Baz test.com nginx</h1>
    <h2>Success!</h2>
    <img
      src="https://em-content.zobj.net/source/microsoft-teams/363/smiling-face-with-heart-eyes_1f60d.png"
      alt="smiling face with heart eyes" />
  </body>
</html>
```

Enable the site configuration by creating a symlink

""

```sh
sudo ln -s /etc/nginx/sites-available/test.com /etc/nginx/sites-enabled/
```

- `ln`: This is the command to create links between files or directories.

- `-s`: This option specifies that you want to create a symbolic link (symlink) instead of a hard link.

"The command is creating a symbolic link from `/etc/nginx/sites-available/test.com` to `/etc/nginx/sites-enabled/test.com`. This effectively "enables" the website configuration for test.com by creating a link in the sites-enabled directory.

By using symlinks, you can easily enable or disable website configurations without actually moving or copying files. You can simply add or remove the symlink in the `sites-enabled` directory to control which configurations are active.

After creating the symlink, you would typically need to restart the Nginx service for the changes to take effect."

Test the nginx configuration for syntax errors

```sh
baz@baz-pc:~$ sudo nginx -t
nginx: the configuration file /etc/nginx/nginx.conf syntax is ok
nginx: configuration file /etc/nginx/nginx.conf test is successful
baz@baz-pc:~$
```

Reload the nginx service to apply the changes

```sh
baz@baz-pc:~$ sudo systemctl reload nginx
baz@baz-pc:~$
```

Visit [http://test.com](http://test.com) but it goes to the real website

"When you're setting up Nginx on your local machine (localhost), you have the flexibility to use any domain name you want, even if you don't own it. This is because you're not actually setting up a publicly accessible website, but rather a local development environment.

In the Nginx configuration, the server_name directive is used to specify the domain name that the server should respond to. When you configure Nginx on localhost, you can use any domain name you want as the server_name, and Nginx will respond to requests for that domain name coming from your local machine.

However, to make this work, you need to map the domain name to the local IP address (`127.0.0.1` or `localhost`) in your local hosts file (`/etc/hosts` on Linux or macOS). The hosts file is used by your operating system to resolve domain names to IP addresses."

Need to overwrite that in the `/etc/hosts` file

```sh
baz@baz-pc:~$ cat /etc/hosts
127.0.0.1 localhost
127.0.1.1 baz-pc

# The following lines are desirable for IPv6 capable hosts
::1     ip6-localhost ip6-loopback
fe00::0 ip6-localnet
ff00::0 ip6-mcastprefix
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters
baz@baz-pc:~$
```

Open the `/etc/hosts` file in VSCode

```sh
baz@baz-pc:~$ code /etc/hosts
```

Add the line:

```sh
127.0.0.1 test.com
```

Reload the nginx service to apply the changes

```sh
baz@baz-pc:~$ sudo systemctl reload nginx
baz@baz-pc:~$
```

Visit [http://test.com](http://test.com) in the browser and I see my own custom nginx page!

![first-nginx-site](images/first-nginx-site.png)
