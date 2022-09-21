package gitaskpass_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/coder/cli/gitaskpass"
)

func TestParse(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Input  string
		Output string
	}{{
		Input:  "Username for 'https://github.com': ",
		Output: "https://github.com",
	}, {
		Input:  "Username for 'https://enterprise.github.com': ",
		Output: "https://enterprise.github.com",
	}, {
		Input:  "Username for 'http://wow.io': ",
		Output: "http://wow.io",
	}} {
		tc := tc
		t.Run(tc.Input, func(t *testing.T) {
			t.Parallel()
			value, err := gitaskpass.Parse(tc.Input)
			require.NoError(t, err)
			require.Equal(t, tc.Output, value)
		})
	}

	// git_origins: []string{"https://github.com"}
}
