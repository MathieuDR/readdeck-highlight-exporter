package util

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func GenerateId(title string) string {
	slug := sluggify(title)
	timestamp := time.Now().Unix()

	return fmt.Sprintf("%d-%s", timestamp, slug)
}

var slugRegex = regexp.MustCompile("[^a-zA-Z0-9]+")

func sluggify(input string) string {
	processedString := slugRegex.ReplaceAllString(input, " ")
	processedString = strings.TrimSpace(processedString)
	slug := strings.ReplaceAll(processedString, " ", "-")

	return strings.ToLower(slug)
}
