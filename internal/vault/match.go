package vault

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

type patternSet [][]string // each pattern pre-split on "/"

func (ps patternSet) match(p string) bool {
	name := strings.Split(p, "/")
	for _, pat := range ps {
		if matchSegs(pat, name) {
			return true
		}
	}
	return false
}

func compile(globs []string, deny bool) (patternSet, error) {
	var ps patternSet
	for _, g := range globs {
		g = path.Clean(filepath.ToSlash(g))
		if _, err := path.Match(g, ""); err != nil {
			return nil, fmt.Errorf("bad glob %q: %w", g, err)
		}
		ps = append(ps, strings.Split(g, "/"))
		if deny && !strings.Contains(g, "**") {
			ps = append(ps, strings.Split(g+"/**", "/"))
		}
	}
	return ps, nil
}

// ** matches zero or more segments; * ? [] match within on segment (no /).
func matchSegs(pat, name []string) bool {
	if len(pat) == 0 {
		return len(name) == 0
	}
	if pat[0] == "**" {
		for i := 0; i < len(name); i++ {
			if matchSegs(pat[1:], name[i:]) {
				return true
			}
		}
		return false
	}
	if len(name) == 0 {
		return false
	}
	if ok, _ := path.Match(pat[0], name[0]); ok {
		return matchSegs(pat[1:], name[1:])
	}
	return false
}
