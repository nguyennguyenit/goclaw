package providers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeCLIProviderChatWithImagesUsesMatchingStreamJSONFormats(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cliPath := filepath.Join(tmpDir, "fake-claude.sh")
	script := `#!/bin/sh
output_format=""
input_format=""

while [ "$#" -gt 0 ]; do
	case "$1" in
		--output-format)
			output_format="$2"
			shift 2
			;;
		--input-format)
			input_format="$2"
			shift 2
			;;
		*)
			shift
			;;
	esac
done

if [ "$input_format" = "stream-json" ] && [ "$output_format" != "stream-json" ]; then
	echo "Error: --input-format=stream-json requires output-format=stream-json." >&2
	exit 1
fi

if [ "$input_format" = "stream-json" ]; then
	cat >/dev/null
fi

printf '%s\n' '{"type":"assistant","message":{"content":[{"type":"text","text":"vision ok"}]}}'
printf '%s\n' '{"type":"result","subtype":"success","result":"vision ok","usage":{"input_tokens":7,"output_tokens":3}}'
`
	if err := os.WriteFile(cliPath, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	p := NewClaudeCLIProvider(cliPath, WithClaudeCLIWorkDir(tmpDir))
	resp, err := p.Chat(context.Background(), ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: "describe this image",
				Images: []ImageContent{
					{MimeType: "image/png", Data: "aGVsbG8="},
				},
			},
		},
		Model: "sonnet",
	})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if resp.Content != "vision ok" {
		t.Fatalf("Chat() content = %q, want %q", resp.Content, "vision ok")
	}
	if resp.Usage == nil {
		t.Fatal("Chat() usage = nil, want usage")
	}
	if resp.Usage.PromptTokens != 7 || resp.Usage.CompletionTokens != 3 || resp.Usage.TotalTokens != 10 {
		t.Fatalf("Chat() usage = %+v, want prompt=7 completion=3 total=10", resp.Usage)
	}
}
