// Package utils provides utility functions.
package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/mitchellh/go-homedir"
)

// RemoveFrontmatter removes the front matter header of a markdown file.
func RemoveFrontmatter(content []byte) []byte {
	if frontmatterBoundaries := detectFrontmatter(content); frontmatterBoundaries[0] == 0 {
		return content[frontmatterBoundaries[1]:]
	}
	return content
}

var yamlPattern = regexp.MustCompile(`(?m)^---\r?\n(\s*\r?\n)?`)

func detectFrontmatter(c []byte) []int {
	if matches := yamlPattern.FindAllIndex(c, 2); len(matches) > 1 {
		return []int{matches[0][0], matches[1][1]}
	}
	return []int{-1, -1}
}

// ExpandPath expands tilde and all environment variables from the given path.
func ExpandPath(path string) string {
	s, err := homedir.Expand(path)
	if err == nil {
		return os.ExpandEnv(s)
	}
	return os.ExpandEnv(path)
}

// WrapCodeBlock wraps a string in a code block with the given language.
func WrapCodeBlock(s, language string) string {
	return "```" + language + "\n" + s + "```"
}

// fencePattern matches the opening or closing line of a fenced code block:
// an optional blockquote prefix, up to three spaces of indentation, the
// fence itself, and the (possibly empty) info string.
var fencePattern = regexp.MustCompile("^((?: {0,3}> ?)* {0,3})(`{3,}|~{3,})(.*)$")

// PlainTextCodeBlocks tags fenced code blocks that don't specify a language
// as plain text. Without a language, chroma guesses one by analyzing the
// block's contents and highlights it accordingly; tagging the fence with
// "text" keeps such blocks unhighlighted while leaving explicitly tagged
// blocks (e.g. ```java) alone. Matching is best-effort per CommonMark;
// fences nested in containers the scanner doesn't model (e.g. list items
// indented 4+ spaces) are left untouched, erring toward not modifying
// content.
func PlainTextCodeBlocks(md string) string {
	var (
		inFence    bool
		fenceChar  byte
		fenceLen   int
		quoteDepth int
	)

	lines := strings.Split(md, "\n")
	for i, line := range lines {
		m := fencePattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		prefix, fence, info := m[1], m[2], strings.TrimSpace(m[3])

		if inFence {
			// Only a matching fence of at least the opening length with no
			// info string closes the block; anything else is content.
			if fence[0] == fenceChar && len(fence) >= fenceLen && info == "" && strings.Count(prefix, ">") == quoteDepth {
				inFence = false
			}
			continue
		}

		// Info strings of backtick fences cannot contain backticks; such a
		// line is not a fence at all (e.g. inline code spans).
		if fence[0] == '`' && strings.Contains(info, "`") {
			continue
		}

		inFence = true
		fenceChar = fence[0]
		fenceLen = len(fence)
		quoteDepth = strings.Count(prefix, ">")

		if info == "" {
			lines[i] = prefix + fence + "text"
		}
	}

	return strings.Join(lines, "\n")
}

var markdownExtensions = []string{
	".md", ".mdown", ".mkdn", ".mkd", ".markdown",
}

// IsMarkdownFile returns whether the filename has a markdown extension.
func IsMarkdownFile(filename string) bool {
	ext := filepath.Ext(filename)

	if ext == "" {
		// By default, assume it's a markdown file.
		return true
	}

	for _, v := range markdownExtensions {
		if strings.EqualFold(ext, v) {
			return true
		}
	}

	// Has an extension but not markdown
	// so assume this is a code file.
	return false
}

// GlamourStyle returns a glamour.TermRendererOption based on the given style.
func GlamourStyle(style string, isCode bool) glamour.TermRendererOption {
	if !isCode {
		if style == styles.AutoStyle {
			return glamour.WithAutoStyle()
		}
		return glamour.WithStylePath(style)
	}

	// If we are rendering a pure code block, we need to modify the style to
	// remove the indentation.

	var styleConfig ansi.StyleConfig

	switch style {
	case styles.AutoStyle:
		if lipgloss.HasDarkBackground() {
			styleConfig = styles.DarkStyleConfig
		} else {
			styleConfig = styles.LightStyleConfig
		}
	case styles.DarkStyle:
		styleConfig = styles.DarkStyleConfig
	case styles.LightStyle:
		styleConfig = styles.LightStyleConfig
	case styles.PinkStyle:
		styleConfig = styles.PinkStyleConfig
	case styles.NoTTYStyle:
		styleConfig = styles.NoTTYStyleConfig
	case styles.DraculaStyle:
		styleConfig = styles.DraculaStyleConfig
	case styles.TokyoNightStyle:
		styleConfig = styles.DraculaStyleConfig
	default:
		return glamour.WithStylesFromJSONFile(style)
	}

	var margin uint
	styleConfig.CodeBlock.Margin = &margin

	return glamour.WithStyles(styleConfig)
}
