package posts

import (
	"html"
	"net/url"
	"strings"
)

func RenderMarkdown(input string) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	lines := strings.Split(input, "\n")
	rendered := make([]string, 0, len(lines))
	for i := 0; i < len(lines); {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			continue
		}
		if strings.HasPrefix(line, "```") {
			htmlBlock, next := renderCodeBlock(lines, i)
			rendered = append(rendered, htmlBlock)
			i = next
			continue
		}
		if isHeading(line) {
			rendered = append(rendered, renderHeading(line))
			i++
			continue
		}
		if isUnorderedListItem(line) {
			htmlBlock, next := renderList(lines, i)
			rendered = append(rendered, htmlBlock)
			i = next
			continue
		}
		if strings.HasPrefix(line, ">") {
			htmlBlock, next := renderQuote(lines, i)
			rendered = append(rendered, htmlBlock)
			i = next
			continue
		}

		htmlBlock, next := renderParagraph(lines, i)
		rendered = append(rendered, htmlBlock)
		i = next
	}

	return strings.Join(rendered, "\n")
}

func renderCodeBlock(lines []string, start int) (string, int) {
	code := make([]string, 0, len(lines)-start)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
			return "<pre><code>" + html.EscapeString(strings.Join(code, "\n")) + "</code></pre>", i + 1
		}
		code = append(code, lines[i])
	}
	return "<pre><code>" + html.EscapeString(strings.Join(code, "\n")) + "</code></pre>", len(lines)
}

func isHeading(line string) bool {
	for level := 6; level >= 1; level-- {
		prefix := strings.Repeat("#", level) + " "
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func renderHeading(line string) string {
	for level := 6; level >= 1; level-- {
		prefix := strings.Repeat("#", level) + " "
		if strings.HasPrefix(line, prefix) {
			return "<h" + itoa(level) + ">" + renderInline(strings.TrimSpace(line[len(prefix):])) + "</h" + itoa(level) + ">"
		}
	}
	return "<p>" + renderInline(line) + "</p>"
}

func isUnorderedListItem(line string) bool {
	return strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ")
}

func renderList(lines []string, start int) (string, int) {
	items := make([]string, 0, len(lines)-start)
	i := start
	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !isUnorderedListItem(line) {
			break
		}
		items = append(items, "<li>"+renderInline(strings.TrimSpace(line[2:]))+"</li>")
	}
	return "<ul>" + strings.Join(items, "") + "</ul>", i
}

func renderQuote(lines []string, start int) (string, int) {
	quoted := make([]string, 0, len(lines)-start)
	i := start
	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, ">") {
			break
		}
		quoted = append(quoted, strings.TrimSpace(strings.TrimPrefix(line, ">")))
	}
	return "<blockquote>" + renderParagraphLines(quoted) + "</blockquote>", i
}

func renderParagraph(lines []string, start int) (string, int) {
	paragraph := make([]string, 0, len(lines)-start)
	i := start
	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "```") || isHeading(line) || isUnorderedListItem(line) || strings.HasPrefix(line, ">") {
			break
		}
		paragraph = append(paragraph, line)
	}
	return renderParagraphLines(paragraph), i
}

func renderParagraphLines(lines []string) string {
	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		rendered = append(rendered, "<p>"+renderInline(strings.TrimSpace(line))+"</p>")
	}
	return strings.Join(rendered, "<br>")
}

func renderInline(input string) string {
	var builder strings.Builder
	for i := 0; i < len(input); {
		if strings.HasPrefix(input[i:], "![") {
			if end := findMarkdownLinkEnd(input, i+2); end > i {
				builder.WriteString(html.EscapeString(input[i : end+1]))
				i = end + 1
				continue
			}
		}
		if strings.HasPrefix(input[i:], "**") {
			if end := strings.Index(input[i+2:], "**"); end >= 0 {
				builder.WriteString("<strong>")
				builder.WriteString(renderInline(input[i+2 : i+2+end]))
				builder.WriteString("</strong>")
				i += end + 4
				continue
			}
		}
		if input[i] == '*' {
			if end := strings.Index(input[i+1:], "*"); end >= 0 {
				builder.WriteString("<em>")
				builder.WriteString(renderInline(input[i+1 : i+1+end]))
				builder.WriteString("</em>")
				i += end + 2
				continue
			}
		}
		if input[i] == '`' {
			if end := strings.Index(input[i+1:], "`"); end >= 0 {
				builder.WriteString("<code>")
				builder.WriteString(html.EscapeString(input[i+1 : i+1+end]))
				builder.WriteString("</code>")
				i += end + 2
				continue
			}
		}
		if input[i] == '[' {
			if closeLabel := strings.Index(input[i+1:], "]("); closeLabel >= 0 {
				labelEnd := i + 1 + closeLabel
				urlStart := labelEnd + 2
				if closeURL := strings.Index(input[urlStart:], ")"); closeURL >= 0 {
					href := input[urlStart : urlStart+closeURL]
					if isSafeHref(href) {
						builder.WriteString("<a href=\"")
						builder.WriteString(html.EscapeString(href))
						builder.WriteString("\" rel=\"nofollow noopener noreferrer\" target=\"_blank\">")
						builder.WriteString(renderInline(input[i+1 : labelEnd]))
						builder.WriteString("</a>")
					} else {
						builder.WriteString(renderInline(input[i+1 : labelEnd]))
					}
					i = urlStart + closeURL + 1
					continue
				}
			}
		}
		builder.WriteString(html.EscapeString(input[i : i+1]))
		i++
	}
	return builder.String()
}

func findMarkdownLinkEnd(input string, labelStart int) int {
	labelEnd := strings.Index(input[labelStart:], "](")
	if labelEnd < 0 {
		return -1
	}
	urlStart := labelStart + labelEnd + 2
	urlEnd := strings.Index(input[urlStart:], ")")
	if urlEnd < 0 {
		return -1
	}
	return urlStart + urlEnd
}

func isSafeHref(href string) bool {
	parsed, err := url.Parse(href)
	if err != nil || parsed.Scheme == "" {
		return false
	}
	scheme := strings.ToLower(parsed.Scheme)
	return scheme == "http" || scheme == "https" || scheme == "mailto"
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
