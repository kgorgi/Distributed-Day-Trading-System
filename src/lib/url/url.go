package url

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

// ResolveAuditServerAddress returns the audit server address
func ResolveAuditServerAddress() string {
	return resolveServerAddress("audit-server", "5002")
}

// ResolveTransactionServerAddress returns the transaction server address
func ResolveTransactionServerAddress() string {
	return resolveServerAddress("transaction-server", "5000")
}

// ResolveDataServerAddress returns the database server address
func ResolveDataServerAddress() string {
	return resolveServerAddress("data-server", "5000")
}

// ResolveMockQuoteServerAddress returns the mock legacy quote server address
func ResolveMockQuoteServerAddress() string {
	return resolveServerAddress("quote-mock-server", "4443")
}

// ResolveDatabaseDBAddress returns the mongo DB address for the database server
func ResolveDatabaseDBAddress() string {
	return resolveMongoAddress("data-mongoDB", "27017", "27017")

}

// ResolveAuditDBAddress returns the mongo DB address for the audit server
func ResolveAuditDBAddress() string {
	return resolveMongoAddress("audit-mongoDB", "27017", "5003")
}
