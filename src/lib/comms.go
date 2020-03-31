package lib

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	security "extremeWorkload.com/daytrader/lib/security"
)

const retries = 3

// In milliseconds
const backoff = 1000

const socketTimeout = time.Duration(60) * time.Second

const seperatorChar = "|"

// StatusOk (HTTP 200)
const StatusOk = 200

// StatusUserError (HTTP 400)
const StatusUserError = 400

// StatusSystemError (HTTP 500)
const StatusSystemError = 500

// StatusNotFound (HTTP 404)
const StatusNotFound = 404

// HealthCheck signal for health check
const HealthCheck = "HEALTH"

// HealthStatusUp healthy status
const HealthStatusUp = "UP"

// HealthStatusTrigger healthy status and trigger
const HealthStatusTrigger = "TRIGGER"

// ClientSendRequest sends a request to a server and then returns
// the response from the server (status, message/error, exception)
func ClientSendRequest(address string, payload string) (int, string, error) {
	return clientSendRequestRetry(address, payload, 1, StatusSystemError, "", nil)
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

// ServerSendHealthResponse sends a healthy response
func ServerSendHealthResponse(conn net.Conn, healthStatus string) error {
	return sendMessage(conn, strconv.Itoa(StatusOk)+seperatorChar+healthStatus)
}

func clientSendRequestRetry(address string, payload string, currentAttempt int, status int, message string, err error) (int, string, error) {
	if currentAttempt >= retries {
		return status, message, err
	}

	if currentAttempt > 1 {
		Debug("Failed request (%d): %s %s\n Err: %s\n", currentAttempt, address, payload, err.Error())
	}

	time.Sleep(time.Duration(currentAttempt-1*backoff) * time.Millisecond)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		currentAttempt++
		return clientSendRequestRetry(address, payload, currentAttempt, StatusSystemError, "", err)
	}

	err = sendMessage(conn, payload)
	if err != nil {
		conn.Close()
		currentAttempt++
		return clientSendRequestRetry(address, payload, currentAttempt, StatusSystemError, "", err)
	}

	respPayload, err := readMessage(conn)
	conn.Close()
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

func sendMessage(conn net.Conn, message string) error {
	conn.SetWriteDeadline(time.Now().Add(socketTimeout))
	encryptedMessage, err := security.Encrypt(message)
	if err != nil {
		return err
	}

	// Create message length header
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(encryptedMessage)))

	// Send header + message
	combined := append(b, encryptedMessage...)
	_, err = conn.Write(combined)
	conn.SetWriteDeadline(time.Time{})
	return err
}

func readMessage(conn net.Conn) (string, error) {
	conn.SetReadDeadline(time.Now().Add(socketTimeout))
	r := bufio.NewReader(conn)

	// Get message length
	b := make([]byte, 8)
	_, err := r.Read(b)
	if err != nil {
		return "", err
	}
	messageLength := int64(binary.LittleEndian.Uint64(b))

	// Get message
	rawPayload := make([]byte, messageLength)
	_, err = io.ReadFull(r, rawPayload)
	if err != nil {
		return "", err
	}
	conn.SetReadDeadline(time.Time{})
	payload, err := security.Decrypt(rawPayload)
	return payload, err
}
