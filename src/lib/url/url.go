package url

import "os"

// IsDocker is the server running in a docker container
func IsDocker() bool {
	value := os.Getenv("IS_DOCKER")
	if len(value) == 0 {
		return false
	}

	return value == "true"
}

func resolveServerAddress(server string, port string) string {
	if IsDocker() {
		return server + ":" + port
	}

	return "localhost:" + port
}

func resolveMongoAddress(server string, port string) string {
	var mongoServer = "localhost"
	if IsDocker() {
		mongoServer = server
	}

	return "mongodb://" + mongoServer + ":" + port + "/mongodb"
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

// ResolveLegacyQuoteServerAddress returns the mock legacy quote server address
func ResolveLegacyQuoteServerAddress() string {
	return resolveServerAddress("quote-mock-server", "4443")
}

// ResolveDatabaseDBAddress returns the mongo DB address for the database server
func ResolveDatabaseDBAddress() string {
	return resolveMongoAddress("data-mongoDB", "27017")

}

// ResolveAuditDBAddress returns the mongo DB address for the audit server
func ResolveAuditDBAddress() string {
	return resolveMongoAddress("audit-mongoDB", "27018")
}
