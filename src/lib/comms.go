package lib

import (
	"bufio"
	"net"
	"strconv"
	"strings"
)

const eotChar = '\x04'
const seperatorChar = "|"

// StatusOk (HTTP 200)
const StatusOk = 200

// StatusUserError (HTTP 400)
const StatusUserError = 400

// StatusSystemError (HTTP 500)
const StatusSystemError = 500

// ClientSendRequest sends a request to a server and then returns
// the response from the server (status, message/error, exception)
func ClientSendRequest(conn net.Conn, payload string) (int, string, error) {
	_, err := conn.Write([]byte(payload + string(eotChar)))
	if err != nil {
		return StatusSystemError, "", err
	}

	rawRespPayload, err := bufio.NewReader(conn).ReadString(eotChar)
	if err != nil {
		return StatusSystemError, "", err
	}

	respPayload := strings.TrimRight(rawRespPayload, string(eotChar))

	data := strings.Split(respPayload, seperatorChar)

	statusCode, err := strconv.Atoi(data[0])
	if err != nil {
		return StatusSystemError, "", err
	}

	if len(data) == 2 {
		return statusCode, data[1], nil
	}

	return statusCode, "", nil
}

// ServerReceiveRequest processes a request from a client
func ServerReceiveRequest(conn net.Conn) (string, error) {
	rawPayload, err := bufio.NewReader(conn).ReadString(eotChar)
	if err != nil {
		return "", err
	}

	payload := strings.TrimRight(rawPayload, string(eotChar))
	return payload, nil
}

// ServerSendOKResponse sends an OK response
func ServerSendOKResponse(conn net.Conn) error {
	payload := strconv.Itoa(StatusOk) + seperatorChar + string(eotChar)
	_, err := conn.Write([]byte(payload))
	return err
}

// ServerSendResponse sends a response to a client
func ServerSendResponse(conn net.Conn, status int, message string) error {
	payload := strconv.Itoa(status) + seperatorChar + message + string(eotChar)
	_, err := conn.Write([]byte(payload))
	return err
}
