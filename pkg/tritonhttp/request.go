package tritonhttp

import (
	"bufio"
	"errors"
	"strings"
)

type Request struct {
	Method string // e.g. "GET"
	URL    string // e.g. "/path/to/a/file"
	Proto  string // e.g. "HTTP/1.1"

	// Header stores misc headers excluding "Host" and "Connection",
	// which are stored in special fields below.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	Header map[string]string

	Host  string // determine from the "Host" header
	Close bool   // determine from the "Connection" header
}

// ReadRequest tries to read the next valid request from br.
//
// If it succeeds, it returns the valid request read. In this case,
// bytesReceived should be true, and err should be nil.
//
// If an error occurs during the reading, it returns the error,
// and a nil request. In this case, bytesReceived indicates whether or not
// some bytes are received before the error occurs. This is useful to determine
// the timeout with partial request received condition.
func ReadRequest(br *bufio.Reader) (req *Request, bytesReceived bool, err error) {
	// panic("todo")
	req = new(Request)
	req.Close = false
	sline, errRead := ReadLine(br)
	if sline == "" {
		bytesReceived = false
		if errRead == nil {
			return nil, bytesReceived, errors.New("No byte Received")
		} else {
			return nil, bytesReceived, errRead
		}

	} else {
		bytesReceived = true
	}
	if errRead != nil {
		err = errRead
		return req, bytesReceived, err
	}

	// Startline
	spli_sline := strings.Split(sline, " ")
	if len(spli_sline) != 3 {
		return req, bytesReceived, errors.New("Wrong Start Line Format")
	}

	req.Method = spli_sline[0]
	if spli_sline[0] != "GET" {
		return req, bytesReceived, errors.New("Method is not GET")
	}

	if strings.HasPrefix(spli_sline[1], "/") {
		// if strings.HasSuffix(spli_sline[1], "/") {
		// 	req.URL = spli_sline[1] + "index.html"
		// } else {
		// 	req.URL = spli_sline[1]
		// }
		req.URL = spli_sline[1]
	} else {
		req.URL = spli_sline[1]
		return req, bytesReceived, errors.New("URL Misses Starting Slash")
	}

	req.Proto = spli_sline[2]
	if spli_sline[2] != "HTTP/1.1" {
		return req, bytesReceived, errors.New("Proto is not HTTP/1.1")
	}

	// Hearders
	req.Header = make(map[string]string)
	hasHost := false
	for {
		l, errRead := ReadLine(br)
		if errRead != nil {
			err = errRead
			break
		}
		if l == "" {
			break
		} else {
			l_split := strings.SplitN(l, ":", 2)
			if len(l_split) != 2 {
				err = errors.New("Hearder Line Wrong Format")
				break
			}
			if checkKey(l_split[0]) {
				l_split[0] = CanonicalHeaderKey(l_split[0])
			} else {
				err = errors.New("Invalid Key")
				break
			}
			// l_split[0] = strings.TrimSpace(l_split[0])
			l_split[1] = strings.TrimLeft(l_split[1], " ")
			if strings.Contains(l_split[1], "\r") {
				err = errors.New("Invalid value")
				break
			}
			if l_split[0] == "Host" {
				req.Host = l_split[1]
				hasHost = true
			} else if l_split[0] == "Connection" {
				// bool_value, errBool := strconv.ParseBool(l_split[1])
				// if errBool != nil {
				// 	err = errBool
				// } else {
				// 	req.Close = bool_value
				// }
				if l_split[1] == "close" {
					req.Close = true
				}
			} else {
				req.Header[l_split[0]] = l_split[1]
			}
		}
	}
	if !hasHost {
		err = errors.New("Missing Host")
	}
	return req, bytesReceived, err

	// Read start line

	// Read headers

	// Check required headers

	// Handle special headers
}

func checkKey(k string) bool {
	if k == "" {
		return false
	}

	for _, c := range k {
		if (c >= '0') && (c <= '9') {
			continue
		} else if (c >= 'a') && (c <= 'z') {
			continue
		} else if (c >= 'A') && (c <= 'Z') {
			continue
		} else if c == '-' {
			continue
		} else {
			return false
		}
	}

	return true

}
