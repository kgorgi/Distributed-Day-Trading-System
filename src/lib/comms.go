package lib

import (
	"bufio"
	"encoding/binary"
	"net"
	"strconv"
	"strings"

	security "extremeWorkload.com/daytrader/lib/security"
)

const seperatorChar = "|"

// StatusOk (HTTP 200)
const StatusOk = 200

// StatusUserError (HTTP 400)
const StatusUserError = 400

// StatusSystemError (HTTP 500)
const StatusSystemError = 500

// StatusNotFound (HTTP 404)
const StatusNotFound = 404

// ClientSendRequest sends a request to a server and then returns
// the response from the server (status, message/error, exception)
func ClientSendRequest(conn net.Conn, payload string) (int, string, error) {
	err := sendMessage(conn, payload)
	if err != nil {
		return StatusSystemError, "", err
	}

	respPayload, err := readMessage(conn)
	if err != nil {
		return StatusSystemError, "", err
	}
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
	return readMessage(conn)
}

// ServerSendOKResponse sends an OK response
func ServerSendOKResponse(conn net.Conn) error {
	return sendMessage(conn, strconv.Itoa(StatusOk)+seperatorChar)
}

// ServerSendResponse sends a response to a client
func ServerSendResponse(conn net.Conn, status int, message string) error {
	return sendMessage(conn, strconv.Itoa(status)+seperatorChar+message)
}

func sendMessage(conn net.Conn, message string) error {
	encryptedMessage := security.Encrypt(message)

	// Create message length header
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(encryptedMessage)))

	// Send header + message
	combined := append(b, encryptedMessage...)
	_, err := conn.Write(combined)
	return err
}

func readMessage(conn net.Conn) (string, error) {
	r := bufio.NewReader(conn)

	// Get message length
	b := make([]byte, 8)
	r.Read(b)
	messageLength := int64(binary.LittleEndian.Uint64(b))

	// Get message
	rawPayload := make([]byte, messageLength)
	_, err := r.Read(rawPayload)
	if err != nil {
		return "", err
	}
	payload := security.Decrypt(rawPayload)
	return payload, nil
}
