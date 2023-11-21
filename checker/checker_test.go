package checker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "gitlab.com/m8127/produtos/automation/dns-checker/checker"
)

func TestRight(t *testing.T) {
	original := "https://google.com"
	target := "https://www.google.com/"
	result, err := Check(original, target, 301)
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestWrong(t *testing.T) {
	for _, tc := range []struct {
		name,
		original,
		target,
		errMsg string
		status int
	}{
		{"Wrong status code", "https://google.com", "https://www.google.com/", "status code: 301", 404},
		{"Wrong target location", "https://google.com", "https://www.akadseguros.com.br/", "destination: https://www.google.com/", 301},
		{"Invalid source url", "not-an-url", "https://www.akadseguros.com.br/", `Get "not-an-url": unsupported protocol scheme ""`, 301},
		{"Invalid target url", "https://google.com", ":not-an-url", `parse ":not-an-url": missing protocol scheme`, 301},
		{"Not a redirect", "https://www.google.com/", "https://www.google.com/", "http: no Location header in response", 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Check(tc.original, tc.target, tc.status)
			assert.EqualError(t, err, tc.errMsg)
			assert.False(t, result)
		})
	}
}
