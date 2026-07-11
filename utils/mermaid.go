package utils

import (
	"os/exec"
	"strings"
)

type mermaidRenderer func(string) (string, error)

// RenderMermaidCodeBlocks replaces Mermaid fenced code blocks with the
// terminal graph produced by mermaid-ascii. Render failures deliberately leave
// the original block unchanged so that unsupported Mermaid syntax remains
// readable.
func RenderMermaidCodeBlocks(md string) string {
	return renderMermaidCodeBlocks(md, renderMermaid)
}

func renderMermaid(input string) (string, error) {
	cmd := exec.Command("mermaid-ascii") //nolint:gosec // Command is not user-controlled.
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func renderMermaidCodeBlocks(md string, render mermaidRenderer) string {
	var (
		out        []string
		block      []string
		inMermaid  bool
		fenceChar  byte
		fenceLen   int
		quoteDepth int
		openLine   string
		prefix     string
	)

	lines := strings.Split(md, "\n")
	for _, line := range lines {
		match := fencePattern.FindStringSubmatch(line)

		if !inMermaid {
			if match == nil {
				out = append(out, line)
				continue
			}

			blockPrefix, fence, info := match[1], match[2], strings.TrimSpace(match[3])
			if fence[0] == '`' && strings.Contains(info, "`") {
				out = append(out, line)
				continue
			}
			if !isMermaidInfoString(info) {
				out = append(out, line)
				continue
			}

			inMermaid = true
			fenceChar = fence[0]
			fenceLen = len(fence)
			quoteDepth = strings.Count(blockPrefix, ">")
			openLine = line
			prefix = blockPrefix
			block = block[:0]
			continue
		}

		if match == nil {
			block = append(block, line)
			continue
		}

		blockPrefix, fence, info := match[1], match[2], strings.TrimSpace(match[3])
		if fence[0] != fenceChar || len(fence) < fenceLen || info != "" || strings.Count(blockPrefix, ">") != quoteDepth {
			block = append(block, line)
			continue
		}

		input := strings.Join(block, "\n")
		if quoteDepth > 0 {
			input = stripBlockQuotePrefixes(block, quoteDepth)
		}
		if rendered, err := render(input); err == nil {
			out = append(out, prefix+fence+"text")
			for _, renderedLine := range strings.Split(strings.TrimSuffix(rendered, "\n"), "\n") {
				out = append(out, prefix+renderedLine)
			}
			out = append(out, blockPrefix+fence)
		} else {
			out = append(out, openLine)
			out = append(out, block...)
			out = append(out, line)
		}

		inMermaid = false
		block = nil
	}

	if inMermaid {
		out = append(out, openLine)
		out = append(out, block...)
	}

	return strings.Join(out, "\n")
}

func isMermaidInfoString(info string) bool {
	fields := strings.Fields(info)
	return len(fields) > 0 && strings.EqualFold(fields[0], "mermaid")
}

func stripBlockQuotePrefixes(lines []string, quoteDepth int) string {
	stripped := make([]string, len(lines))
	for i, original := range lines {
		line := original
		for range quoteDepth {
			line = strings.TrimPrefix(line, " ")
			line = strings.TrimPrefix(line, " ")
			line = strings.TrimPrefix(line, " ")
			if !strings.HasPrefix(line, ">") {
				break
			}
			line = strings.TrimPrefix(line, ">")
			line = strings.TrimPrefix(line, " ")
		}
		stripped[i] = line
	}
	return strings.Join(stripped, "\n")
}
