package serverurls

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// ServerUrls store all the urls needed to communicate between servers
type ServerUrls struct {
	WebServer         string `yaml:"web"`
	LegacyQuoteServer string `yaml:"legacyQuoteServer"`
	AuditServer       string `yaml:"audit"`
	DataServer        string `yaml:"data"`
	TransactionServer string `yaml:"transaction"`
	QuoteCacheServer  string `yaml:"quoteCache"`
	AuditDBServer     string `yaml:"auditDB"`
	DataDBServer      string `yaml:"dataDB"`
}

const defaultConfigFilePath = "urls.yml"

func getUrlsConfig() ServerUrls {
	var urlFilePath = os.Getenv("URLS_FILE")
	if urlFilePath == "" {
		urlFilePath = defaultConfigFilePath
	}

	urls := make(map[string]ServerUrls)

	yamlFile, err := ioutil.ReadFile(urlFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	err = yaml.Unmarshal(yamlFile, &urls)
	if err != nil {
		log.Fatalln(err)
	}

	var env = os.Getenv("ENV")

	switch env {
	case "DOCKER":
		return urls["docker"]
	case "DEV":
		return urls["dev"]
	case "DEV-LAB":
		return urls["dev-lab"]
	case "LAB":
		return urls["lab"]
	default:
		return urls["local"]
	}
}

// Env used to access the current execution environment Urls
var Env = getUrlsConfig()
