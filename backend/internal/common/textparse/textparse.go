package textparse

import (
	"regexp"
	"strings"
)

var (
	topicPattern   = regexp.MustCompile(`(?:^|\s)#([\p{L}\p{N}_-]{1,40})`)
	mentionPattern = regexp.MustCompile(`(?:^|[^\p{L}\p{N}_])@([A-Za-z0-9_]{1,32})`)
)

func ExtractTopics(input string) []string {
	matches := topicPattern.FindAllStringSubmatch(input, -1)
	seen := make(map[string]struct{}, len(matches))
	topics := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		topic := strings.ToLower(strings.Trim(match[1], "_-"))
		if topic == "" {
			continue
		}
		if _, ok := seen[topic]; ok {
			continue
		}
		seen[topic] = struct{}{}
		topics = append(topics, topic)
	}
	return topics
}

func ExtractMentions(input string) []string {
	matches := mentionPattern.FindAllStringSubmatch(input, -1)
	seen := make(map[string]struct{}, len(matches))
	mentions := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		username := strings.TrimSpace(match[1])
		if username == "" {
			continue
		}
		key := strings.ToLower(username)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		mentions = append(mentions, username)
	}
	return mentions
}
