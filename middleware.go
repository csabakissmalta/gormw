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

// requestID -> originalToken
var sessionIDs map[string][]string

// var cntr int = 0

func main() {
	sessionIDs = make(map[string][]string)

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
			ret := v
			return ret
		}
	}
	return ""
}

func process(buf []byte) {
	payloadType := buf[0]
	headerSize := bytes.IndexByte(buf, '\n') + 1
	// header := buf[:headerSize-1]

	// Header contains space separated values of: request type, request id, and request start time (or round-trip time for responses)
	// meta := bytes.Split(header, []byte(" "))
	// reqID := string(meta[1])
	payload := buf[headerSize:]

	switch payloadType {
	case '1':
		os.Stdout.Write(encode(buf))
	case '2':
		Debug("---- THIS IS TURBOLOGIN ORIG RESPONSE ----")
		hs := proto.ParseHeaders(payload)
		for key, ele := range hs {
			if key == "Set-Cookie" {
				resp := get_session_id(ele)
				if len(resp) > 0 {
					sessionIDs[resp] = nil
				}
			}
		}
	case '3':
		hs := proto.ParseHeaders(payload)
		for key, ele := range hs {
			if key == "Set-Cookie" {
				resp := get_session_id(ele)
				if len(resp) > 0 {
					if value, ok := sessionIDs[resp]; ok {
						fmt.Println("value: ", value)
					} else {
						fmt.Println("key not found")
					}
				}
			}
		}
		Debug("---------------- REPLAY ----------------")
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
