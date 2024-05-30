package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Request struct {
	Method   string // GET, POST etc
	Path     string
	HTTP     string
	Headers  map[string]string
	Contents []byte
}

type Response struct {
	Code     int
	Status   string
	Headers  map[string]string
	Contents string
}

func parseUntil(request []byte, start int, delimiter string) (int, []byte) {
	var output []byte
	end := []byte(delimiter)

	i := start
	b := request[i : i+len(end)]

	for i < len(request) && !bytes.Equal(b, end) {
		output = append(output, b[0])
		i += 1
		b = request[i : i+len(end)]
	}
	fmt.Println("Parsed: --" + string(output) + "--")

	val := i + len(delimiter)
	return val, output
}

func parseRequest(request []byte) Request {
	fmt.Println("Parsing: --" + string(request) + "--")

	i, method := parseUntil(request, 0, " ")
	i, path := parseUntil(request, i, " ")
	i, http := parseUntil(request, i, "\r\n")

	headers := make(map[string]string)
	for request[i] != '\r' && request[i+1] != '\n' {
		var name, value []byte
		i, name = parseUntil(request, i, ": ")
		i, value = parseUntil(request, i, "\r\n")
		headers[string(name)] = string(value)
		fmt.Println("Header: --" + string(name) + ": " + string(value) + "--")
	}
	i += 2 // skip the final \r\n

	i, contents := parseUntil(request, i, "\x00")
	fmt.Printf("Contents: %q\n", string(contents))

	return Request{
		Method:   string(method),
		Path:     string(path),
		HTTP:     string(http),
		Headers:  headers,
		Contents: contents,
	}
}

func (r *Response) build() string {

	response := "HTTP/1.1 "
	response += strconv.Itoa(r.Code) + " "
	response += r.Status + "\r\n"

	for key, value := range r.Headers {
		response += key + ": " + value + "\r\n"
	}
	if len(r.Contents) > 0 {
		response += "Content-Length: " + strconv.Itoa(len(r.Contents)) + "\r\n"
	}
	response += "\r\n"
	response += r.Contents

	fmt.Printf("Response: \n --- \n%q\n", response)

	return response
}

func validEncoding(encoding string) string {
	split := strings.Split(encoding, ", ")
	for _, s := range split {
		if s == "gzip" {
			return s
		}
	}
	return ""
}

func gzipCompress(data string) string {
	fmt.Printf("Compressing %q\n", data)
	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)
	i, e := w.Write([]byte(data))
	w.Close()
	if e != nil {
		fmt.Println(e)
	}
	bytes := buf.Bytes()
	fmt.Printf("Compressed %d bytes %q\n", i, bytes)

	// encoded := make([]byte, hex.EncodedLen(len(bytes)))
	// hex.Encode(encoded, bytes)
	// fmt.Printf("Resulting encoding: %d %q\n", len(encoded), encoded)

	return string(bytes)
}

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading HTTP request header")
	}
	request := parseRequest(buffer)

	// response := "HTTP/1.1 404 Not Found\r\n\r\n"
	response := Response{
		Code:     404,
		Status:   "Not Found",
		Headers:  map[string]string{},
		Contents: "",
	}
	if request.Method == "GET" {
		if strings.HasPrefix(request.Path, "/echo") {
			str := request.Path[6:]
			fmt.Println("Extracted string: --" + str + "--")
			response.Code = 200
			response.Status = "OK"
			response.Headers["Content-Type"] = "text/plain"
			response.Contents = str

			encoding, exists := request.Headers["Accept-Encoding"]
			if exists && len(validEncoding(encoding)) > 0 {
				encoded := gzipCompress(str)
				response.Headers["Content-Encoding"] = validEncoding(encoding)
				response.Contents = encoded
			}
		} else if strings.HasPrefix(request.Path, "/user-agent") {
			response.Code = 200
			response.Status = "OK"
			response.Headers["Content-Type"] = "text/plain"
			response.Contents = request.Headers["User-Agent"]
		} else if strings.HasPrefix(request.Path, "/files") {
			directory := os.Args[2]
			filepath := directory + "/" + request.Path[7:]
			fmt.Printf("Checking existence of file %q\n", filepath)
			_, err := os.Stat(filepath)
			if err == nil {
				b, _ := os.ReadFile(filepath)
				contents := string(b)
				response.Code = 200
				response.Status = "OK"
				response.Headers["Content-Type"] = "application/octet-stream"
				response.Contents = contents
			} // else defaults to 404
		} else if request.Path == "/" {
			response.Code = 200
			response.Status = "OK"
		}
	} else if request.Method == "POST" {
		if strings.HasPrefix(request.Path, "/files") {
			directory := os.Args[2]
			filepath := directory + "/" + request.Path[7:]
			os.WriteFile(filepath, request.Contents, 0666)
			if err != nil {
				fmt.Println(err)
			}
			response.Code = 201
			response.Status = "Created"
		}
	}

	_, err = conn.Write([]byte(response.build()))
	if err != nil {
		fmt.Println("Error writing HTTP header")
		os.Exit(1)
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
		// conn.Close()
	}

}
