package reporting

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindEventIndex(t *testing.T) {
	// Test case 1: Normal case with event found
	eventsData := [][]string{
		{"Launch", "0.0", "0.0", "0.0"},
		{"Burnout", "1.5", "100.0", "50.0"},
		{"Apogee", "5.0", "500.0", "0.0"},
		{"Touchdown", "35.0", "0.0", "-5.0"},
	}
	
	idx := FindEventIndex(eventsData, "Apogee")
	assert.Equal(t, 2, idx, "Should find the correct event index for 'Apogee'")
	
	// Test case 2: Case insensitive match
	idx = FindEventIndex(eventsData, "apogee")
	assert.Equal(t, 2, idx, "Should find the event index in a case-insensitive manner")
	
	idx = FindEventIndex(eventsData, "BURNOUT")
	assert.Equal(t, 1, idx, "Should find the event index in a case-insensitive manner")
	
	// Test case 3: Event not found
	idx = FindEventIndex(eventsData, "NonExistentEvent")
	assert.Equal(t, -1, idx, "Should return -1 when event is not found")
	
	// Test case 4: Nil events data
	idx = FindEventIndex(nil, "Launch")
	assert.Equal(t, -1, idx, "Should return -1 when events data is nil")
	
	// Test case 5: Empty events data
	idx = FindEventIndex([][]string{}, "Launch")
	assert.Equal(t, -1, idx, "Should return -1 when events data is empty")
	
	// Test case 6: Trailing/leading whitespace in the data (not in the input)
	eventsWithWhitespace := [][]string{
		{"Launch", "0.0", "0.0", "0.0"},
		{" Burnout ", "1.5", "100.0", "50.0"}, // Note the extra whitespace
	}
	
	idx = FindEventIndex(eventsWithWhitespace, "Burnout")
	assert.Equal(t, 1, idx, "Should handle whitespace in the event data")
}

func TestConvertMarkdownToSimpleHTML(t *testing.T) {
	// Test case 1: Basic conversion with headers and paragraphs
	mdContent := "# Main Header\n\n## Second Level Header\n\n### Third Level Header\n\nThis is a paragraph.\n\nThis is another paragraph."
	html := ConvertMarkdownToSimpleHTML(mdContent, "test-record-id")
	
	// Check that the HTML contains expected elements
	assert.Contains(t, html, "<h1>Simulation Report: test-record-id</h1>", "Should include report title with record ID")
	assert.Contains(t, html, "<h1>Main Header</h1>", "Should convert # to h1")
	assert.Contains(t, html, "<h2>Second Level Header</h2>", "Should convert ## to h2")
	assert.Contains(t, html, "<h3>Third Level Header</h3>", "Should convert ### to h3")
	assert.Contains(t, html, "<p>This is a paragraph.</p>", "Should convert paragraphs")
	assert.Contains(t, html, "<p>This is another paragraph.</p>", "Should convert multiple paragraphs")
	
	// Test case 2: Code blocks
	mdContent = "```\nfunction test() {\n  console.log('Hello');\n}\n```"
	html = ConvertMarkdownToSimpleHTML(mdContent, "test-record-id")
	assert.Contains(t, html, "<pre><code>function test() {\n  console.log('Hello');\n}</code></pre>", "Should handle code blocks")
	
	// Test case 3: HTML escaping of < character
	mdContent = "# Header with <script>\n\nParagraph with <strong>HTML</strong>."
	html = ConvertMarkdownToSimpleHTML(mdContent, "test-record-id")
	// Check for actual content with the exact same string format as in the implementation
	assert.Contains(t, html, "<h1>Header with &lt;script></h1>", "Should escape < characters in headers")
	assert.Contains(t, html, "<p>Paragraph with &lt;strong>HTML&lt;/strong>.</p>", "Should escape < characters in paragraphs")
	
	// Test case 4: Verify HTML document structure
	mdContent = "# Simple Test"
	html = ConvertMarkdownToSimpleHTML(mdContent, "test-record-id")
	assert.True(t, strings.HasPrefix(html, "<!DOCTYPE html>"), "Should start with DOCTYPE declaration")
	assert.Contains(t, html, "<html>", "Should include html tag")
	assert.Contains(t, html, "<head>", "Should include head tag")
	assert.Contains(t, html, "<style>", "Should include CSS styles")
	assert.Contains(t, html, "<body>", "Should include body tag")
	assert.True(t, strings.HasSuffix(html, "</html>"), "Should end with closing HTML tag")
}
