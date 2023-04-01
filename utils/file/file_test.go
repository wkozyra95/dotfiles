package file

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	content              string
	text                 string
	rg                   *regexp.Regexp
	resultShouldUpdate   bool
	resultUpdatedContent string
}

func TestEnsureLineText(t *testing.T) {
	cases := []testCase{
		{
			content:              "export CURRENT_ENV=\"home\"\n",
			text:                 "export CURRENT_ENV=\"work\"",
			rg:                   regexp.MustCompile("export CURRENT_ENV.*"),
			resultShouldUpdate:   true,
			resultUpdatedContent: "export CURRENT_ENV=\"work\"\n",
		},
		{
			content:              " ",
			text:                 "export CURRENT_ENV=\"work\"",
			rg:                   regexp.MustCompile("export CURRENT_ENV.*"),
			resultShouldUpdate:   true,
			resultUpdatedContent: " \nexport CURRENT_ENV=\"work\"\n",
		},
		{
			content:              "test",
			text:                 "export CURRENT_ENV=\"work\"",
			rg:                   regexp.MustCompile("export CURRENT_ENV.*"),
			resultShouldUpdate:   true,
			resultUpdatedContent: "test\nexport CURRENT_ENV=\"work\"\n",
		},
		{
			content:              "test\n",
			text:                 "export CURRENT_ENV=\"work\"",
			rg:                   regexp.MustCompile("export CURRENT_ENV.*"),
			resultShouldUpdate:   true,
			resultUpdatedContent: "test\nexport CURRENT_ENV=\"work\"\n",
		},
		{
			content:              "test\nexport CURRENT_ENV=\"home\"\ntest2\n",
			text:                 "export CURRENT_ENV=\"work\"",
			rg:                   regexp.MustCompile("export CURRENT_ENV.*"),
			resultShouldUpdate:   true,
			resultUpdatedContent: "test\nexport CURRENT_ENV=\"work\"\ntest2\n",
		},
		{
			content:              "test\n          export CURRENT_ENV=\"home\"\ntest2\n",
			text:                 "export CURRENT_ENV=\"work\"",
			rg:                   regexp.MustCompile("[ \t]*export CURRENT_ENV.*"),
			resultShouldUpdate:   true,
			resultUpdatedContent: "test\nexport CURRENT_ENV=\"work\"\ntest2\n",
		},
		{
			content:              "test\nexport CURRENT_ENV=\"home\"\ntest2\n",
			text:                 "export CURRENT_ENV=\"home\"",
			rg:                   regexp.MustCompile("export CURRENT_ENV.*"),
			resultShouldUpdate:   false,
			resultUpdatedContent: "test\nexport CURRENT_ENV=\"home\"\ntest2\n",
		},
		{
			content:              "aaabbbccc",
			text:                 "d",
			rg:                   regexp.MustCompile("bbb"),
			resultShouldUpdate:   true,
			resultUpdatedContent: "aaadccc",
		},
		{
			content:              "aaaccc",
			text:                 "ddd",
			rg:                   regexp.MustCompile(regexp.QuoteMeta("ddd")),
			resultShouldUpdate:   true,
			resultUpdatedContent: "aaaccc\nddd\n",
		},
		{
			content:              "aaadddccc",
			text:                 "ddd",
			rg:                   regexp.MustCompile(regexp.QuoteMeta("ddd")),
			resultShouldUpdate:   false,
			resultUpdatedContent: "aaadddccc",
		},
	}

	for i, testCase := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			shouldUpdate, content := ensureTextInString(testCase.content, testCase.text, testCase.rg)
			assert.Equal(t, testCase.resultShouldUpdate, shouldUpdate)
			assert.Equal(t, testCase.resultUpdatedContent, content)
		})
	}
}
