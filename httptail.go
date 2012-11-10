// Copyright (C) 2012 Chen "smallfish" Xiaoyu (陈小玉)
package main

import (
    "flag"
    "fmt"
    "io"
    "net"
    "net/http"
    "os"
    "strings"
)

var mode string  // server or client
var redis string // redis address, default: 0.0.0.0:6379
var topic string // when mode equal client, publish to redis topic, default: default
var bind string  // when mode equal server, bind httpserver host:port, default: 0.0.0.0:8888

var redisPubProtocol = "*3\r\n$7\r\npublish\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n"
var redisSubProtocol = "*2\r\n$9\r\nsubscribe\r\n$%d\r\n%s\r\n"

func init() {
    flag.StringVar(&mode, "mode", "", "server or client")
    flag.StringVar(&redis, "redis", "0.0.0.0:6379", "redis host:port")
    flag.StringVar(&topic, "topic", "default", "publish topic")
    flag.StringVar(&bind, "bind", "0.0.0.0:8888", "bind httpserver host:port")
}

func serverModeHandler(redis, bind string) {
    // create default handler
    http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
        conn, err := net.Dial("tcp", redis)
        if err != nil {
            fmt.Println("error: connect redis failed.", err)
            return
        }
        defer conn.Close()
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        io.WriteString(w, "httptail start...\n")
        w.(http.Flusher).Flush()
        topic := req.URL.Path[1:]
        _, err = conn.Write([]byte(fmt.Sprintf(redisSubProtocol, len(topic), topic)))
        if err != nil {
            fmt.Println("error: subscribe redis failed.", err)
            return
        }
        var buf = make([]byte, 1024)
        for {
            n, err := conn.Read(buf)
            if err == io.EOF {
                break
            }
            if n > 0 {
                lines := strings.Split(string(buf[0:n]), "\r\n")
                array := make([]string, 0)
                for i := 0; i < len(lines); i++ {
                    if strings.Index(lines[i], "$") == 0 { // ignore *
                        array = append(array, lines[i+1])
                    }
                }
                if array[0] == "message" { // ignore subscribe
                    io.WriteString(w, fmt.Sprintf("%s %s", array[1], array[2]))
                    w.(http.Flusher).Flush()
                }
            }
        }
    })

    // server bind
    err := http.ListenAndServe(bind, nil)
    if err != nil {
        fmt.Println("error: server bind failed.", err)
    }
}

func clientModeHandler(redis, topic string) {
    conn, err := net.Dial("tcp", redis)
    if err != nil {
        fmt.Println("error: connect redis failed.", err)
        return
    }
    defer conn.Close()
    buf := make([]byte, 1024)
    for {
        n, err := os.Stdin.Read(buf)
        if err == io.EOF {
            break
        }
        if n > 0 {
            value := string(buf[0:n])
            command := fmt.Sprintf(redisPubProtocol, len(topic), topic, len(value), value)
            _, err := conn.Write([]byte(command))
            if err != nil {
                fmt.Println("error: publish to redis failed.", err)
                return
            }
            fmt.Print(value)
        }
    }
}

func main() {
    flag.Parse()
    if mode != "server" && mode != "client" { // check mode
        fmt.Println("error: mode must server or client")
        return
    }
    if mode == "client" { // client mode
        clientModeHandler(redis, topic)
        return
    }
    if mode == "server" { // server mode
        serverModeHandler(redis, bind)
        return
    }
}
