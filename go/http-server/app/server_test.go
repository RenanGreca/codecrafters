package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestParseUntil(t *testing.T) {
	header := "GET / HTTP/1.1"
	_, output := parseUntil([]byte(header), 0, " ")
	if !bytes.Equal(output, []byte("GET")) {
		t.Fatalf(`Expected %q, received %q`, string(output), "GET")
	}
}

func TestParseUntilTwo(t *testing.T) {
	header := "HTTP/1.1\r\nContent-Type: text/plain"
	_, output := parseUntil([]byte(header), 0, "\r\n")
	if !bytes.Equal(output, []byte("HTTP/1.1")) {
		t.Fatalf(`Expected %q, received %q`, string(output), "HTTP/1.1")
	}
}

func TestParseUntilConsecutive(t *testing.T) {
	header := "GET / HTTP/1.1"
	i, output := parseUntil([]byte(header), 0, " ")
	fmt.Println(i)
	if !bytes.Equal(output, []byte("GET")) {
		t.Fatalf(`Expected %q, received %q`, string(output), "GET")
	}
	i, output = parseUntil([]byte(header), i, " ")
	fmt.Println(i)
	if !bytes.Equal(output, []byte("/")) {
		t.Fatalf(`Expected %q, received %q`, string(output), "/")
	}
}
