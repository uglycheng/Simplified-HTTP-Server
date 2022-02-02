package tritonhttp

import (
	"errors"
	"io"
	"os"
	"sort"
	"strconv"
)

type Response struct {
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.1"

	// Header stores all headers to write to the response.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	Header map[string]string

	// Request is the valid request that leads to this response.
	// It could be nil for responses not resulting from a valid request.
	Request *Request

	// FilePath is the local path to the file to serve.
	// It could be "", which means there is no file to serve.
	FilePath string
}

// Write writes the res to the w.
func (res *Response) Write(w io.Writer) error {
	if err := res.WriteStatusLine(w); err != nil {
		return err
	}
	if err := res.WriteSortedHeaders(w); err != nil {
		return err
	}
	if err := res.WriteBody(w); err != nil {
		return err
	}
	return nil
}

// WriteStatusLine writes the status line of res to w, including the ending "\r\n".
// For example, it could write "HTTP/1.1 200 OK\r\n".
func (res *Response) WriteStatusLine(w io.Writer) error {
	// panic("todo")
	if res.Proto == "HTTP/1.1" {
		w.Write([]byte("HTTP/1.1 "))
	} else {
		return errors.New("Invalid Response Proto: " + res.Proto)
	}

	statusMap := map[int]string{
		200: "OK",
		400: "Bad Request",
		404: "Not Found",
	}
	statusDes, ok := statusMap[res.StatusCode]
	if !ok {
		return errors.New("Invalid Status Code")
	}
	w.Write([]byte(strconv.Itoa(res.StatusCode)))
	w.Write([]byte(" " + statusDes + "\r\n"))

	return nil
}

// WriteSortedHeaders writes the headers of res to w, including the ending "\r\n".
// For example, it could write "Connection: close\r\nDate: foobar\r\n\r\n".
// For HTTP, there is no need to write headers in any particular order.
// TritonHTTP requires to write in sorted order for the ease of testing.
func (res *Response) WriteSortedHeaders(w io.Writer) error {
	keys := make([]string, 0, len(res.Header))
	for k, _ := range res.Header {
		// canKey := CanonicalHeaderKey(k)
		// res.Header[canKey] = v
		// keys = append(keys, canKey)
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, kOrder := range keys {
		w.Write([]byte(kOrder + ": " + res.Header[kOrder] + "\r\n"))
	}
	w.Write([]byte("\r\n"))
	return nil
}

// WriteBody writes res' file content as the response body to w.
// It doesn't write anything if there is no file to serve.
func (res *Response) WriteBody(w io.Writer) error {
	if res.FilePath != "" {
		var bytesWant []byte
		var err error
		bytesWant, err = os.ReadFile(res.FilePath)
		if err != nil {
			return errors.New("Can not Read File")
		}
		w.Write(bytesWant)
	}
	return nil
}
