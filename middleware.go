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
var originalSessionIDs map[string][]string

// originalToken -> replayedToken
var replayedSessionIDs map[string][]string

// var cntr int = 0

func main() {
	originalSessionIDs = make(map[string][]string)
	replayedSessionIDs = make(map[string][]string)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		encoded := scanner.Bytes()
		buf := make([]byte, len(encoded)/2)
		hex.Decode(buf, encoded)

		process(buf)
	}
}

// func read_old_session_data(payload []byte) {

// }

func process(buf []byte) {
	// First byte indicate payload type, possible values:
	//  1 - Request
	//  2 - Response
	//  3 - ReplayedResponse
	payloadType := buf[0]
	headerSize := bytes.IndexByte(buf, '\n') + 1
	// header := buf[:headerSize-1]

	// Header contains space separated values of: request type, request id, and request start time (or round-trip time for responses)
	// meta := bytes.Split(header, []byte(" "))
	// reqID := string(meta[1])
	payload := buf[headerSize:]

	switch payloadType {
	case '1':
		req_path := proto.Path(payload)
		if strings.Contains(string(req_path), "turboLogin") {
			hs := proto.ParseHeaders(payload)
			for key, ele := range hs {
				if key == "Set-Cookie" {
					Debug("---- Found the old cookies ----", ele)
				}
				Debug(">> REPLAY ", string(key), ele)
			}
			// Debug("<< REQ PATH", string(req_path))
			os.Stdout.Write(encode(buf))
		}

	case '2':
		Debug("---- THIS IS TURBOLOGIN ORIG RESPONSE ----")
		os.Stdout.Write(encode(buf))
		// hs := proto.ParseHeaders(payload)
		// for key, ele := range hs {
		// 	Debug(">> ORIG RESP:", string(key), ele)
		// }
	case '3':
		// stat := proto.Status(payload)
		hs := proto.ParseHeaders(payload)
		for key, ele := range hs {
			Debug(">> REPLAY ", string(key), ele)
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
