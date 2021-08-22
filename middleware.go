package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/csabakissmalta/gormw/proto"
)

type old_to_new struct {
	old string
	new []string
}

// requestID -> originalToken
var sessionIDs map[string]old_to_new

// var cntr int = 0

func main() {
	sessionIDs = make(map[string]old_to_new)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		encoded := scanner.Bytes()
		buf := make([]byte, len(encoded)/2)
		hex.Decode(buf, encoded)

		process(buf)
	}
}

func get_session_id(ele []string) string {
	for _, v := range ele {
		if strings.Contains(v, "SESSION_ID") {
			// Debug("::: SESSION_ID", v)
			clean_raw_v := strings.Split(v, ";")[0]
			return clean_raw_v
		}
	}
	return ""
}

func get_session_id_from_cookie(ele []string) string {
	for _, v := range ele {
		if strings.Contains(v, "SESSION_ID") {
			// Debug("::: SESSION_ID", v)
			for _, val := range strings.Split(v, ";") {
				if strings.Contains(val, "SESSION_ID") {
					return val
				}
			}

		}
	}
	return ""
}

func create_cookie_value_from_list(lst []string) string {
	raw := *new([]string)
	for _, v := range lst {
		// append(raw, strings.Split(v, ";")[0])
		raw = append(raw, strings.Split(v, ";")[0])
	}
	res := strings.Join(raw, ";")
	return res
}

func process(buf []byte) {
	payloadType := buf[0]
	headerSize := bytes.IndexByte(buf, '\n') + 1
	header := buf[:headerSize-1]

	// Header contains space separated values of: request type, request id, and request start time (or round-trip time for responses)
	meta := bytes.Split(header, []byte(" "))
	reqID := string(meta[1])
	payload := buf[headerSize:]

	switch payloadType {
	case '1':
		hs := proto.ParseHeaders(payload)
		for key, ele := range hs {
			if key == "Cookie" {
				resp := get_session_id_from_cookie(ele)
				// Debug(string(resp))
				if len(resp) > 4 {
					// if value, ok := sessionIDs[string(resp)]; ok {
					// 	// set the new header
					// 	new_cookie := create_cookie_value_from_list(value)
					// 	Debug("--- NC: ", new_cookie)
					// 	proto.SetHeader(payload, []byte("Cookie"), []byte(new_cookie))
					// }
					for _, val := range sessionIDs {
						Debug(val.old)
						if val.old == resp {

							new_cookie := create_cookie_value_from_list(val.new)
							Debug("--- NC: ", new_cookie)
							proto.SetHeader(payload, []byte("Cookie"), []byte(new_cookie))
						}
					}
				}
			}
		}
		os.Stdout.Write(encode(buf))
	case '2':
		// Debug("---- THIS IS TURBOLOGIN ORIG RESPONSE ----")
		hs := proto.ParseHeaders(payload)
		for key, ele := range hs {
			if key == "Set-Cookie" {
				resp := get_session_id(ele)
				if len(resp) > 4 {
					// Debug(string(resp))
					sessionIDs[reqID] = old_to_new{old: resp}
				}
			}
		}
	case '3':
		hs := proto.ParseHeaders(payload)
		for key, ele := range hs {
			if key == "Set-Cookie" {
				resp := get_session_id(ele)
				// Debug("--- GETTING NEW COOKIE: ", resp)
				if len(resp) > 4 {
					if value, ok := sessionIDs[reqID]; ok {
						value.new = ele
					}
				}
			}
		}
		// Debug(":: Status: ", string(proto.Status(payload)))
	}
}

// --------------------------------------------------------------------------
func Debug(args ...interface{}) {
	if os.Getenv("GOR_TEST") == "" {
		fmt.Fprint(os.Stderr, "[DEBUG][TOKEN-MOD] ")
		fmt.Fprintln(os.Stderr, args...)
	}
}

func encode(buf []byte) []byte {
	dst := make([]byte, len(buf)*2+1)
	hex.Encode(dst, buf)
	dst[len(dst)-1] = '\n'

	return dst
}
