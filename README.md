socker2stdout
==========

Listen on a TCP or Unix socket and send to stdout.

Why?
====

This was written to handle wanting to write to stdout from inside a container from a PHP
application. However, [php-fpm](https://bugs.php.net/bug.php?id=71880) prefixes each line with a preamble.  So, this allows one to configure PHP loggers - such as [monolog](https://github.com/Seldaek/monolog) - to log
to a tcp socket and still [capture stdout](https://12factor.net/logs).

socket2stdout handles log lines greater than PIPE_BUF. Some tools may interleave
lines with large messages.

Usage
=====

```
$ ./socket2stdout -h
copy lines from a socket to stdout

Usage:
  socket2stdout [flags]

Flags:
      --addr string       tcp address (default "127.0.0.1:4444")
      --aux-addr string   listen address for aux handler. metrics, healthchecks, etc (default ":9090")
      --unix string       unix address. takes precedence if both unix and tcp are set
```

An example using monolog to log to this socket:

```php
$logger = new Monolog\Logger("example");
$handler = new Monolog\Handler\SocketHandler('tcp://127.0.0.1:1313');
$formatter = new Monolog\Formatter\JsonFormatter();
$handler->setFormatter($formatter);
$logger->pushHandler($handler);
```

Building
========

```
go build .
````

Also availible at https://quay.io/repository/bakins/socket2stdout
