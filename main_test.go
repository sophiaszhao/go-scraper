package main

import (
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/stretchr/testify/assert"
)

func TestCountHeadings(t *testing.T) {
	html := `
	<html>
	<body>
	<h1>hello</h1>
	<h2>world</h2>
	<h2>a second heading</h2>
	</body>
	</html>`

	doc, err := htmlquery.Parse(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	headerCounts, err := countHeadings(doc)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, headerCounts["h1"])
	assert.Equal(t, 2, headerCounts["h2"])
}
