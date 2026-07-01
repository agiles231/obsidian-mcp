package vault

import "testing"

func TestMatch(t *testing.T) {
	data := []struct {
		p    patternSet
		path string
		exp  bool
	}{
		{[][]string{[]string{"**", "b"}}, "a/b", true},
		{[][]string{[]string{"a", "**"}}, "a/b", true},
		{[][]string{[]string{"**", "a", "**", "b"}}, "a/b", true},
		{[][]string{[]string{"*", "*"}}, "a/b", true},
		{[][]string{[]string{"a", "*"}}, "a/b", true},
		{[][]string{[]string{"*", "b"}}, "a/b", true},
		{[][]string{[]string{"a", "b"}}, "a/b", true},
		{[][]string{[]string{"**"}}, "a/b", true},
		{[][]string{[]string{"a", "**", "b"}}, "a/b", true},
		{[][]string{[]string{"a", "b", "c"}}, "a/b/c", true},
		{[][]string{[]string{"a", "*", "c"}}, "a/b/c", true},
		{[][]string{[]string{"a", "*", "b"}}, "a/b", false},
		{[][]string{[]string{"x", "b"}}, "a/b", false},
		{[][]string{[]string{"*", "b"}}, "b", false},
	}

	for _, d := range data {
		res := d.p.match(d.path)
		if res != d.exp {
			t.Errorf("pattern %v, path %v, expected %v; got %v", d.p, d.path, d.exp, res)
		}
	}

}
