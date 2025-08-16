package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Respond struct {
	body string
	headers map[string]string
	endpoint *string
}

func (r *Request)handlerHome(conn net.Conn) {
	res := &Respond{
		body: "",
		headers: make(map[string]string),
	}
	respond(res, conn, r, nil)
}

func (r *Request)handlerUserAgent(conn net.Conn) {
	res := &Respond{}
	userAgentValue, ok := r.headers["User-Agent"]
	if !ok {
		respond(res, conn, r, errors.New("couldn't find User-Agent header"))
	}
	res.headers = map[string]string {
		"Content-Type": "text/plain",
		"Content-Length": strconv.Itoa(len(userAgentValue)),
	}
	res.body = userAgentValue
	respond(res, conn, r, nil)
}

func (r *Request)handlerEcho(conn net.Conn) {
	acceptEncodingVals := strings.Split(r.headers["Accept-Encoding"], ",")
	gzip := r.compressions[0]
	compressedGzip, err := compressGzip(r.pathParameters)
	res := &Respond{}
	if err != nil {
		log.Printf("Couldn't compress a string\nError: %v\n", err)
		respond(res, conn, r, err)
		return
	}

	if slices.Contains(acceptEncodingVals, gzip) {
		res.headers = map[string]string {
			"Content-Type": "text/plain",
			"Content-Length": strconv.Itoa(len(compressedGzip)),
			"Content-Encoding": gzip,
		}
		res.body = string(compressedGzip)
		respond(res, conn, r, nil)
		return
	}

	res.headers = map[string]string {
		"Content-Type": "text/plain",
		"Content-Length": strconv.Itoa(len(r.pathParameters)),
	}
	res.body = r.pathParameters
	respond(res, conn, r, nil)

}

func (r *Request)handlerFiles(conn net.Conn) {
	fileName := r.pathParameters
	directory := os.Args[2]
	fullFilePath :=  directory + fileName
	res := &Respond{}

	if r.method == "POST" {
		file, err := os.Create(fullFilePath)
		if err != nil {
			log.Printf("Couldn't create a file\nError: %v\n", err)
			respond(res, conn, r, err)
		}
		if _, err := file.Write([]byte(r.body)); err != nil {
			log.Printf("Couldn't write to %v file\nError: %v\n", file.Name(), err)
			respond(res, conn, r, err)
		}
		resEndpoint := "HTTP/1.1 201 Created\r\n"
		res.endpoint = &resEndpoint
		respond(res, conn, r, nil)
		return
	}

	file, err := os.ReadFile(fullFilePath)
	if err != nil {
		log.Printf("The file %v doesn't exist\nError: %v\n", fileName, err)
		respond(res, conn, r, err)
		return
	}

	res.headers = map[string]string{
		"Content-Type": "application/octet-stream",
		"Content-Length": strconv.Itoa(len(file)),
	}
	res.body = string(file)
	respond(res, conn, r, nil)
}

func compressGzip(body string) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)

	if _, err := gzWriter.Write([]byte(body)); err != nil {
		return nil, err
	}
	if err := gzWriter.Close(); err != nil {
		return nil, err
	}
	return  buf.Bytes(), nil
}

func respond(res *Respond, conn net.Conn, req *Request, err error) {
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
