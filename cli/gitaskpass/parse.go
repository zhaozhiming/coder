package gitaskpass

import (
	"regexp"
	"strings"

	"golang.org/x/xerrors"
)

var (
	hostReplace = regexp.MustCompile(`^["']+|["':]+$`)
)

// Parse returns the host from a git ask pass prompt.
func Parse(prompt string) (string, string, error) {
	parts := strings.Split(prompt, " ")
	if len(parts) < 3 {
		return "", "", xerrors.Errorf("got <3 parts: %q", prompt)
	}
	// https://github.com/microsoft/vscode/blob/328646ebc2f5016a1c67e0b23a0734bd598ec5a8/extensions/git/src/askpass-main.ts#L41
	host := parts[2]
	host = hostReplace.ReplaceAllString(host, "")
	return parts[0], host, nil
}
