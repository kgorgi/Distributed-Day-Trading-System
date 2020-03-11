package resolveurl

import "os"

// dockerHost translates to the IP address of local computer
const dockerHost = "host.docker.internal"
const localhost = "localhost"

const auditHost = "192.168.1.200"

var env = os.Getenv("ENV")

func resolveServerAddress(dockerServer string, port string) string {
	var server string

	switch env {
	case "LAB":
		server = dockerServer
	case "DOCKER":
		server = dockerServer
	case "DEV":
		server = dockerHost
	default:
		// Running server locally
		server = localhost
	}

	return server + ":" + port
}

func resolveMongoAddress(dockerServer string, dockerPort string, localPort string) string {
	var server string
	var port string

	switch env {
	case "LAB":
		server = dockerServer
		port = dockerPort
	case "DOCKER":
		server = dockerServer
		port = dockerPort
	case "DEV":
		server = dockerHost
		port = localPort
	default:
		// Running server locally
		server = localhost
		port = localPort
	}

	return "mongodb://" + server + ":" + port
}

// AuditServerAddress returns the audit server address
func resolveAuditServerAddress() string {
	var server string

	switch env {
	case "LAB":
		server = auditHost
	case "DOCKER":
		server = "audit-server"
	case "DEV":
		server = dockerHost
	default:
		// Running server locally
		server = localhost
	}

	return server + ":" + "5002"
}

// AuditServerAddress returns the audit server address
var AuditServerAddress = resolveAuditServerAddress()

// TransactionServerAddress returns the transaction server address
var TransactionServerAddress = resolveServerAddress("transaction-server", "5000")

// DataServerAddress returns the database server address
var DataServerAddress = resolveServerAddress("data-server", "5001")

// MockQuoteServerAddress returns the mock legacy quote server address
var MockQuoteServerAddress = resolveServerAddress("quote-mock-server", "4443")

// DatabaseDBAddress returns the mongo DB address for the database server
var DatabaseDBAddress = resolveMongoAddress("data-mongodb", "27017", "27017")

// AuditDBAddress returns the mongo DB address for the audit server
var AuditDBAddress = resolveMongoAddress("audit-mongodb", "27017", "5003")
