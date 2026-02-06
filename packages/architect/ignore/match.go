// Package ignore...(TODO)
package ignore

import (
	"path/filepath"
	"strings"
)

type pattern struct {
	raw      string
	segments []string
	negate   bool
	dirOnly  bool
	anchored bool
	hasStar2 bool
	invalid  bool
}

func Match(patterns []string, path string, isDir bool) bool {
	parsed := make([]pattern, 0, len(patterns))
	for _, raw := range patterns {
		if p, ok := parsePattern(raw); ok {
			parsed = append(parsed, p)
		}
	}

	ignored := false

	for _, p := range parsed {
		if p.invalid {
			continue
		}

		if !matchPattern(p, path, isDir) {
			continue
		}

		if p.negate {
			if ignored && isParentExcluded(parsed, path) {
				continue
			}
			ignored = false
		} else {
			ignored = true
		}
	}

	return ignored
}

func isParentExcluded(patterns []pattern, path string) bool {
	parts := strings.Split(path, "/")
	if len(parts) <= 1 {
		return false
	}

	for depth := 1; depth < len(parts); depth++ {
		parent := strings.Join(parts[:depth], "/")
		excluded := false
		for _, p := range patterns {
			if p.invalid {
				continue
			}
			if !matchPattern(p, parent, true) {
				continue
			}
			if p.negate {
				excluded = false
			} else {
				excluded = true
			}
		}
		if excluded {
			return true
		}
	}
	return false
}

func parsePattern(raw string) (pattern, bool) {
	var p pattern
	p.raw = raw

	if raw == "" {
		return p, false
	}
	if strings.HasPrefix(raw, "#") {
		return p, false
	}

	line := stripTrailingSpaces(raw)
	if line == "" {
		return p, false
	}

	if strings.HasPrefix(line, "!") {
		p.negate = true
		line = line[1:]
	} else if strings.HasPrefix(line, "\\!") {
		line = "!" + line[2:]
	}

	if strings.HasPrefix(line, "\\#") {
		line = "#" + line[2:]
	}

	if strings.HasSuffix(line, "/") {
		p.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}

	if strings.HasPrefix(line, "/") {
		p.anchored = true
		line = line[1:]
	} else if strings.Contains(line, "/") {
		p.anchored = true
	}

	if isTrailingBackslash(line) {
		p.invalid = true
		return p, true
	}

	p.hasStar2 = strings.Contains(line, "**")
	p.segments = strings.Split(line, "/")
	return p, true
}

func isTrailingBackslash(s string) bool {
	if s == "" || s[len(s)-1] != '\\' {
		return false
	}
	n := 0
	for i := len(s) - 1; i >= 0 && s[i] == '\\'; i-- {
		n++
	}
	return n%2 == 1
}

func stripTrailingSpaces(s string) string {
	end := len(s)
	for end > 0 && s[end-1] == ' ' {
		bs := 0
		for j := end - 2; j >= 0 && s[j] == '\\'; j-- {
			bs++
		}
		if bs%2 == 1 {
			break
		}
		end--
	}
	return s[:end]
}

func matchPattern(p pattern, path string, isDir bool) bool {
	if p.dirOnly && !isDir {
		return false
	}

	pathParts := strings.Split(path, "/")

	if p.hasStar2 {
		return matchDoubleStar(p.segments, pathParts)
	}

	if p.anchored {
		return matchSegments(p.segments, pathParts)
	}

	if len(p.segments) == 1 {
		return matchGlob(p.segments[0], pathParts[len(pathParts)-1])
	}

	for i := 0; i <= len(pathParts)-len(p.segments); i++ {
		if matchSegments(p.segments, pathParts[i:i+len(p.segments)]) {
			return true
		}
	}
	return false
}

func matchDoubleStar(pat, path []string) bool {
	return matchDSRec(pat, path, 0, 0)
}

func matchDSRec(pat, path []string, pi, qi int) bool {
	for pi < len(pat) && qi < len(path) {
		seg := pat[pi]

		if seg == "**" {
			if pi == len(pat)-1 {
				return true
			}
			for k := qi; k <= len(path); k++ {
				if matchDSRec(pat, path, pi+1, k) {
					return true
				}
			}
			return false
		}

		if !matchGlob(seg, path[qi]) {
			return false
		}
		pi++
		qi++
	}

	for pi < len(pat) && pat[pi] == "**" {
		pi++
	}

	return pi == len(pat) && qi == len(path)
}

func matchSegments(patSegs, pathParts []string) bool {
	if len(patSegs) != len(pathParts) {
		return false
	}
	for i := range patSegs {
		if !matchGlob(patSegs[i], pathParts[i]) {
			return false
		}
	}
	return true
}

func matchGlob(pat, name string) bool {
	ok, err := filepath.Match(pat, name)
	if err != nil {
		return false
	}
	return ok
}
