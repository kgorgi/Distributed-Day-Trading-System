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
	TransactionServer string `yaml:"transaction"`
	QuoteCacheServer  string `yaml:"quoteCache"`
	AuditDBServer     string `yaml:"auditDB"`
	DataDBServer      string `yaml:"dataDB"`
}

// Urls contains the parsed urls in urls.yml
type Urls struct {
	Serve map[string]ServerUrls `yaml:"urls"`
	Watch map[string][]string   `yaml:"watchdog"`
}

const defaultConfigFilePath = "urls.yml"

// GetUrlsConfig Obtains the data from urls.yml
func GetUrlsConfig() Urls {
	var urlFilePath = os.Getenv("URLS_FILE")
	if urlFilePath == "" {
		urlFilePath = defaultConfigFilePath
	}

	var urls Urls

	yamlFile, err := ioutil.ReadFile(urlFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	err = yaml.Unmarshal(yamlFile, &urls)
	if err != nil {
		log.Fatalln(err)
	}
	return urls
}

func getServerUrls() ServerUrls {
	urls := GetUrlsConfig()
	var env = os.Getenv("ENV")

	switch env {
	case "DOCKER":
		return urls.Serve["docker"]
	case "DEV":
		return urls.Serve["dev"]
	case "DEV-LAB":
		return urls.Serve["dev-lab"]
	case "LAB":
		return urls.Serve["lab"]
	default:
		return urls.Serve["local"]
	}
}

// Env used to access the current execution environment Urls
var Env = getServerUrls()
