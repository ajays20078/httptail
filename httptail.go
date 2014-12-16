// Copyright (C) 2012 Chen "smallfish" Xiaoyu (陈小玉)
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var mode string  // server or client
var redis string // redis address, default: 0.0.0.0:6379
var topic string // when mode equal client, publish to redis topic, default: default
var bind string  // when mode equal server, bind httpserver host:port, default: 0.0.0.0:8888

var redisPubProtocol = "*3\r\n$7\r\npublish\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n"
var redisSubProtocol = "*2\r\n$9\r\nsubscribe\r\n$%d\r\n%s\r\n"

type RedisConn struct {
	conn   net.Conn
	buffer bufio.ReadWriter
}

func getRedisConn(address string) (conn *RedisConn, err error) {
	var nc net.Conn
	nc, err = net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &RedisConn{conn: nc, buffer: bufio.ReadWriter{bufio.NewReader(nc), bufio.NewWriter(nc)}}, nil
}

func init() {
	flag.StringVar(&mode, "mode", "", "server or client")
	flag.StringVar(&redis, "redis", "0.0.0.0:6379", "redis host:port")
	flag.StringVar(&topic, "topic", "default", "publish topic")
	flag.StringVar(&bind, "bind", "0.0.0.0:8888", "bind httpserver host:port")
}

func serverModeHandler(redis, bind string) {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		rc, err := getRedisConn(redis)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		defer rc.conn.Close()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, "httptail start...\n")
		w.(http.Flusher).Flush()
		topic := req.URL.Path[1:]
		command := fmt.Sprintf(redisSubProtocol, len(topic), topic)
		if _, err := rc.buffer.WriteString(command); err != nil {
			fmt.Println("error: subscribe redis failed.", err)
			return
		}
		rc.buffer.Flush()
		for {
			line, prefix, err := rc.buffer.ReadLine()
			if prefix || err != nil {
				fmt.Println("error:", err)
				return
			}
			resp := string(line)
			if strings.HasPrefix(resp, "*") {
				bulk, _ := strconv.ParseInt(resp[1:], 10, 0)
				row := make([]string, 0)
				for i := 0; i < int(bulk); i++ {
					line, _, _ := rc.buffer.ReadLine()
					resp := string(line)
					if strings.HasPrefix(resp, "$") {
						count, _ := strconv.ParseInt(resp[1:], 10, 0)
						buf := make([]byte, count+2)
						io.ReadFull(rc.buffer, buf)
						row = append(row, string(buf[:count]))
					}
				}
				io.WriteString(w, strings.Join(row[1:], " "))
				io.WriteString(w, "<br/>")
				w.(http.Flusher).Flush()
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
	rc, err := getRedisConn(redis)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer rc.conn.Close()
	buf := make([]byte, 1024)
	for {
		n, err := os.Stdin.Read(buf)
		if err == io.EOF {
			break
		}
		if n > 0 {
			value := string(buf[0:n])
			command := fmt.Sprintf(redisPubProtocol, len(topic), topic, len(value), value)
			if _, err := rc.buffer.WriteString(command); err != nil {
				fmt.Println("error: publish redis failed.", err)
				return
			}
			rc.buffer.Flush()
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
