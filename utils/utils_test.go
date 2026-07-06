package utils

import "testing"

func TestPlainTextCodeBlocks(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "bare fence gets tagged",
			in:   "# Title\n\n```\nfunc main() {}\n```\n",
			want: "# Title\n\n```text\nfunc main() {}\n```\n",
		},
		{
			name: "explicit language untouched",
			in:   "```java\nclass Foo {}\n```\n",
			want: "```java\nclass Foo {}\n```\n",
		},
		{
			name: "content lines inside fence untouched",
			in:   "```\n```java is a fence\n```\n",
			want: "```text\n```java is a fence\n```\n",
		},
		{
			name: "tilde fence gets tagged",
			in:   "~~~\ncode\n~~~\n",
			want: "~~~text\ncode\n~~~\n",
		},
		{
			name: "backticks inside tilde fence untouched",
			in:   "~~~\n```\ncode\n```\n~~~\n",
			want: "~~~text\n```\ncode\n```\n~~~\n",
		},
		{
			name: "longer fence inside shorter is a close",
			in:   "```\ncode\n````\n",
			want: "```text\ncode\n````\n",
		},
		{
			name: "shorter fence inside longer is content",
			in:   "````\n```\ncode\n````\n",
			want: "````text\n```\ncode\n````\n",
		},
		{
			name: "indented fence up to three spaces",
			in:   "   ```\ncode\n   ```\n",
			want: "   ```text\ncode\n   ```\n",
		},
		{
			name: "four spaces is an indented code block, not a fence",
			in:   "    ```\n    code\n",
			want: "    ```\n    code\n",
		},
		{
			name: "blockquoted bare fence gets tagged",
			in:   "> ```\n> code\n> ```\n",
			want: "> ```text\n> code\n> ```\n",
		},
		{
			name: "backtick fence with backtick in info is not a fence",
			in:   "``` `foo` ```\n",
			want: "``` `foo` ```\n",
		},
		{
			name: "fence with trailing whitespace only info",
			in:   "```   \ncode\n```\n",
			want: "```text\ncode\n```\n",
		},
		{
			name: "multiple blocks handled independently",
			in:   "```\nplain\n```\n\n```go\npackage main\n```\n\n```\nplain again\n```\n",
			want: "```text\nplain\n```\n\n```go\npackage main\n```\n\n```text\nplain again\n```\n",
		},
		{
			name: "idempotent",
			in:   "```text\nplain\n```\n",
			want: "```text\nplain\n```\n",
		},
		{
			name: "unclosed fence",
			in:   "```\ncode\n",
			want: "```text\ncode\n",
		},
		{
			name: "no code blocks",
			in:   "just some *markdown*\n",
			want: "just some *markdown*\n",
		},
		{
			name: "blockquoted fence line inside a fence is content",
			in:   "```\n> ```\n```\nAfter\n",
			want: "```text\n> ```\n```\nAfter\n",
		},
		{
			name: "bare fence does not close a blockquoted fence",
			in:   "> ```\n> code\n\n```\nplain\n```\n",
			want: "> ```text\n> code\n\n```\nplain\n```\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := PlainTextCodeBlocks(tc.in); got != tc.want {
				t.Errorf("PlainTextCodeBlocks(%q)\n got: %q\nwant: %q", tc.in, got, tc.want)
			}
		})
	}
}
