package utils

import "strings"

func ExtractSubject(data []byte) string {
	lines := strings.Split(string(data), "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.ToUpper(line), "SUBJECT:") {
			return strings.TrimSpace(line[len("Subject:"):])
		}
	}
	return "(No Subject)"
}