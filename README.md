Name
====

httptail - tools push stdout/stderr to http chunked

Require
====

    * Golang (test 1.0.3)
    * Redis  (test 2.6.x)

Install
====
    $ go get github.com/smallfish/httptail

Usage
====

    $ ./httptail --help
    Usage of ./httptail:
      -bind="0.0.0.0:8888": bind httpserver host:port
      -mode="": server or client
      -redis="0.0.0.0:6379": redis host:port
      -topic="default": publish topic

Example
====

    1. start server
        $ ./httptail -mode=server -bind="127.0.0.1:9999" -redis="127.0.0.1:6379"

    2. start client
        $ python -u test.py | ./httptail -mode=client -redis="127.0.0.1:6379" -topic=default
        
    3. curl test
        $ curl http://127.0.0.1:9999/default
        httptail start...
