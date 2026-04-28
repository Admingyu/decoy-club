package posts

import (
	"strings"
	"testing"
)

func TestRenderMarkdownSupportsRichPasteOutput(t *testing.T) {
	html := RenderMarkdown(strings.Join([]string{
		"## Launch notes",
		"",
		"Hello **Decoy**, read [the docs](https://example.com/path?q=1).",
		"",
		"- First item",
		"- *Second* item",
	}, "\n"))

	expected := strings.Join([]string{
		"<h2>Launch notes</h2>",
		"<p>Hello <strong>Decoy</strong>, read <a href=\"https://example.com/path?q=1\" rel=\"nofollow noopener noreferrer\" target=\"_blank\">the docs</a>.</p>",
		"<ul><li>First item</li><li><em>Second</em> item</li></ul>",
	}, "\n")
	if html != expected {
		t.Fatalf("expected rendered rich paste markdown\n%s\ngot\n%s", expected, html)
	}
}
