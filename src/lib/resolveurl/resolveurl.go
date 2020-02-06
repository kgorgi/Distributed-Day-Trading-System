package resolveurl

import "os"

// dockerHost translates to the IP address of local computer
const dockerHost = "host.docker.internal"
const localhost = "localhost"

func resolveServerAddress(dockerServer string, port string) string {
	var server string

	env := os.Getenv("ENV")
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

	env := os.Getenv("ENV")

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
func AuditServerAddress() string {
	return resolveServerAddress("audit-server", "5002")
}

// TransactionServerAddress returns the transaction server address
func TransactionServerAddress() string {
	return resolveServerAddress("transaction-server", "5000")
}

// DataServerAddress returns the database server address
func DataServerAddress() string {
	return resolveServerAddress("data-server", "5001")
}

// MockQuoteServerAddress returns the mock legacy quote server address
func MockQuoteServerAddress() string {
	return resolveServerAddress("quote-mock-server", "4443")
}

// DatabaseDBAddress returns the mongo DB address for the database server
func DatabaseDBAddress() string {
	return resolveMongoAddress("data-mongodb", "27017", "27017")

}

// AuditDBAddress returns the mongo DB address for the audit server
func AuditDBAddress() string {
	return resolveMongoAddress("audit-mongodb", "27017", "5003")
}
