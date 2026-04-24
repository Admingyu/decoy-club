package posts

import (
	"html"
	"strings"
)

func RenderMarkdown(input string) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	blocks := strings.Split(input, "\n\n")
	rendered := make([]string, 0, len(blocks))
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		rendered = append(rendered, renderBlock(block))
	}

	return strings.Join(rendered, "\n")
}

func renderBlock(block string) string {
	lines := strings.Split(block, "\n")
	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		rendered = append(rendered, renderLine(line))
	}
	return strings.Join(rendered, "<br>")
}

func renderLine(line string) string {
	trimmed := strings.TrimSpace(line)
	for level := 6; level >= 1; level-- {
		prefix := strings.Repeat("#", level) + " "
		if strings.HasPrefix(trimmed, prefix) {
			return "<h" + itoa(level) + ">" + html.EscapeString(strings.TrimSpace(trimmed[len(prefix):])) + "</h" + itoa(level) + ">"
		}
	}

	return "<p>" + html.EscapeString(trimmed) + "</p>"
}

func itoa(v int) string {
	if v == 1 {
		return "1"
	}
	if v == 2 {
		return "2"
	}
	if v == 3 {
		return "3"
	}
	if v == 4 {
		return "4"
	}
	if v == 5 {
		return "5"
	}
	return "6"
}
