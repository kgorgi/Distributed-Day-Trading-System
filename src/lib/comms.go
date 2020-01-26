package lib

import (
	"bufio"
	"net"
	"strings"
)

const eotChar = '\x04'
const seperatorChar = "|"

// StatusOk (HTTP 200)
const StatusOk = "OK"

// StatusUserERROR (HTTP 400)
const StatusUserERROR = "USER_ERROR"

// StatusSystemError (HTTP 500)
const StatusSystemError = "SYSTEM_ERROR"

// ClientSendRequest sends a request to a server and then returns
// the response from the server (status, message/error, exception)
func ClientSendRequest(conn net.Conn, payload string) (string, string, error) {
	_, err := conn.Write([]byte(payload + string('\x04')))
	if err != nil {
		return "", "", err
	}

	rawRespPayload, err := bufio.NewReader(conn).ReadString(eotChar)
	if err != nil {
		return "", "", err
	}

	respPayload := strings.TrimRight(rawRespPayload, string(eotChar))

	data := strings.Split(respPayload, seperatorChar)
	if len(data) == 2 {
		return data[0], data[1], nil
	}

	return data[0], "", nil
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
	payload := StatusOk + seperatorChar + string(eotChar)
	_, err := conn.Write([]byte(payload))
	return err
}

// ServerSendResponse sends a response to a client
func ServerSendResponse(conn net.Conn, status string, message string) error {
	payload := status + seperatorChar + message + string(eotChar)
	_, err := conn.Write([]byte(payload))
	return err
}
