package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

type Request struct {
	path string
	method string
	body string
	headers map[string]string
	compressions []string
	pathParameters string
}

// handleConnection func handles persistent connection 
func handleConnection(conn net.Conn) {
	for {
		req, err := readParseRequest(conn)
		if err != nil {
			break
		}
		handleRequest(req, conn)
	}
}

func getHeaderVal(header string) (string, string) {
	headerVal := strings.Split(header, ":")
	return headerVal[0], headerVal[1]
}
// readParseRequest func reads the request and returns a Request struct
func readParseRequest(conn net.Conn) (*Request, error) {
	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer);
	if err != nil {
		return nil, fmt.Errorf("error accepting connection: %v", err)
	}
	req := string(buffer[:n])
	lines := strings.Split(req, CRLF)

	// setting the request
	newRequest := &Request{
		path: strings.Split(lines[0], " ")[1],
		method: strings.Split(lines[0], " ")[0],
		body: lines[len(lines) - 1],
		headers: make(map[string]string),
		compressions: []string{"gzip"},
	}
	for _, header := range lines[1:len(lines) - 2] {
		head, val := getHeaderVal(header)
		newRequest.headers[head] = strings.ReplaceAll(val, " ", "")
	}

	return newRequest, nil
}

// handleRequest func handles requests endpoint
func handleRequest(req *Request, conn net.Conn) {
	if req.path == "/" {
		req.handlerHome(conn)

	} else if req.path == "/user-agent" {
		req.handlerUserAgent(conn)

	} else if pathStr, ok := strings.CutPrefix(req.path, "/echo/"); ok {
		req.pathParameters = pathStr
		req.handlerEcho(conn)

	} else if fileName, ok := strings.CutPrefix(req.path, "/files/"); ok {
		req.pathParameters = fileName
		req.handlerFiles(conn)
		
	} else {
		res := &Respond{}
		res.write(conn, req, errors.New("method doesn't exist "))
	}
}

func (res *Respond) write(conn net.Conn, req *Request, err error) {
	var fullRes string
	resErr := "HTTP/1.1 404 Not Found\r\n\r\n"
	resOk := "HTTP/1.1 200 OK\r\n"

	if err != nil {
		fullRes += resErr
		conn.Write([]byte(fullRes))
		return

	} else if res.endpoint == nil {
		res.endpoint = &resOk
	}

	fullRes += *res.endpoint
	headerVal, ok := req.headers["Connection"]
	if ok && headerVal == "close" {
		res.headers["Connection"] = "close"
		defer conn.Close()
	}
	for header, val := range res.headers {
		fullRes += fmt.Sprintf("%s: %s\r\n", header, val)
	}
	fullRes += fmt.Sprintf("\r\n%s", res.body)
	conn.Write([]byte(fullRes))
}
