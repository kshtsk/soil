package util

import (
	"testing"
)
import "github.com/stretchr/testify/assert"

type G struct {
	Ref string
	Url string
	Bra string
}

func TestRepoFromRef(t *testing.T) {
	cases := []G{
		{
			Ref: "https://github.com/user/repo@branch",
			Url: "https://github.com/user/repo",
			Bra: "branch",
		},
		{
			Ref: "https://git-hub.com/user.name/repo.name@br-0",
			Url: "https://git-hub.com/user.name/repo.name",
			Bra: "br-0",
		},
		{
			Ref: "https://git_hub.com/user_name/repo-name.git",
			Url: "https://git_hub.com/user_name/repo-name.git",
			Bra: "",
		},
		{
			Ref: "https://git_hub.com/user_name/repo-name.git@refs/heads/main",
			Url: "https://git_hub.com/user_name/repo-name.git",
			Bra: "refs/heads/main",
		},
	}
	for _, c := range cases {
		url, bra := RepoFromRef(c.Ref)
		assert.Equal(t, c.Url, url)
		assert.Equal(t, c.Bra, bra)
	}
}
