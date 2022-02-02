package tritonhttp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	// Addr specifies the TCP address for the server to listen on,
	// in the form "host:port". It shall be passed to net.Listen()
	// during ListenAndServe().
	Addr string // e.g. ":0"

	// DocRoot specifies the path to the directory to serve static files from.
	DocRoot string
}

// ListenAndServe listens on the TCP network address s.Addr and then
// handles requests on incoming connections.
func (s *Server) ListenAndServe() error {
	if errRoot := s.ValidateServerSetup(); errRoot != nil {
		return errRoot
	}

	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		panic("Can not create a listenner")
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			panic("Accept Error")
		}
		go s.HandleConnection(conn)
	}

	// Hint: call HandleConnection
}

// HandleConnection reads requests from the accepted conn and handles them.
func (s *Server) HandleConnection(conn net.Conn) {
	br := bufio.NewReader(conn)
	// Hint: use the other methods below

	for {
		// Set timeout
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		req, bytesReceived, err := ReadRequest(br)
		if errors.Is(err, io.EOF) {
			conn.Close()
			return
		}

		if (os.IsTimeout(err)) && (!bytesReceived) {
			conn.Close()
			return
		}
		res := new(Response)
		if err != nil {
			res.HandleBadRequest()
		} else {
			res = s.HandleGoodRequest(req)
		}
		var buffer bytes.Buffer
		errSend := res.Write(&buffer)
		if errSend != nil {
			panic(errSend)
		}
		// panic(buffer)
		conn.Write(buffer.Bytes())
		if req.Close || res.StatusCode == 400 {
			conn.Close()
			return
		}

	}
}

// HandleGoodRequest handles the valid req and generates the corresponding res.
func (s *Server) HandleGoodRequest(req *Request) (res *Response) {
	res = new(Response)
	absDocRoot, errPath := filepath.Abs(filepath.Clean(s.DocRoot))
	if errPath != nil {
		panic("DocRootError")
	}
	if strings.HasSuffix(req.URL, "/") {
		req.URL = filepath.Join(req.URL, "index.html")
	}
	path := s.DocRoot + req.URL
	path, errPath = filepath.Abs(filepath.Clean(path))
	if errPath != nil {
		res.HandleNotFound(req)
		return res
	}

	if strings.HasPrefix(path, absDocRoot) {
		_, errPath = os.Stat(path)
		if errPath != nil {
			res.HandleNotFound(req)
		} else {
			res.HandleOK(req, path)
			// rel_path := path[len(absDocRoot):]
			// // panic(rel_path)
			// res.FilePath = rel_path
		}
	} else {
		res.HandleNotFound(req)
	}
	return res
	// Hint: use the other methods below
}

// HandleOK prepares res to be a 200 OK response
// ready to be written back to client.
func (res *Response) HandleOK(req *Request, path string) {
	fileInfo, _ := os.Stat(path)
	res.StatusCode = 200
	res.Proto = "HTTP/1.1"
	res.Header = make(map[string]string)
	if req.Close {
		res.Header["Connection"] = "close"
	}
	res.Header["Date"] = FormatTime(time.Now())
	res.Header["Last-Modified"] = FormatTime(fileInfo.ModTime())
	res.Header["Content-Length"] = strconv.FormatInt(fileInfo.Size(), 10)
	// ext_list := strings.Split(path, ".")
	// ext := "." + ext_list[len(ext_list)-1]
	res.Header["Content-Type"] = MIMETypeByExtension(filepath.Ext(path))
	res.FilePath = path
	res.Request = req
}

// HandleBadRequest prepares res to be a 400 Bad Request response
// ready to be written back to client.
func (res *Response) HandleBadRequest() {
	res.StatusCode = 400
	res.Proto = "HTTP/1.1"
	res.Header = make(map[string]string)
	res.Header["Connection"] = "close"
	res.Header["Date"] = FormatTime(time.Now())
}

// HandleNotFound prepares res to be a 404 Not Found response
// ready to be written back to client.
func (res *Response) HandleNotFound(req *Request) {
	res.StatusCode = 404
	res.Proto = "HTTP/1.1"
	res.Header = make(map[string]string)
	res.Header["Date"] = FormatTime(time.Now())
	if req.Close {
		res.Header["Connection"] = "close"
	}
	res.Request = req
}

func (s *Server) ValidateServerSetup() error {
	// Validating the doc root of the server
	absRoot, errAbs := filepath.Abs(s.DocRoot)
	if errAbs != nil {
		return errors.New("Invalid Root")
	}
	fi, err := os.Stat(absRoot)

	if os.IsNotExist(err) {
		return err
	}

	if !fi.IsDir() {
		return errors.New("Root is not a directory")
	}

	return nil
}

// func (s *Server) Test() string {
// 	a := filepath.Clean(s.DocRoot)
// 	return a
// }
