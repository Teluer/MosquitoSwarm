package util

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
)

var userAgents []string

func RandomUserAgent() string {
	const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/113.0"

	if len(userAgents) == 0 {
		readUserAgents()
	}

	valuesCount := len(userAgents)
	if valuesCount == 0 {
		log.Error("Failed to read user-agents from file, falling back to the default user-agent")
		return defaultUserAgent
	}

	randomIndex := rand.Intn(valuesCount)
	return userAgents[randomIndex]
}

func readUserAgents() {
	file, err := os.Open("user-agents")
	defer file.Close()
	if err != nil {
		log.WithError(err).Error("Failed to open 'user-agents' file!")
		return
	}

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		userAgents = append(userAgents, fileScanner.Text())
	}
}
