package main

import (
	"bufio"
	"os"
	"strings"
)

const privacyVersionPrefix = "Last updated: "

func parsePrivacyVersion(path string) (string, error) { // #nosec G304 -- path is from config, not user input
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, privacyVersionPrefix) {
			return strings.TrimPrefix(line, privacyVersionPrefix), nil
		}
	}

	return "", scanner.Err()
}
