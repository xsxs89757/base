package main

import (
	"regexp"
	"strconv"
)

// 只认正式版 vX.Y.Z。基底不打预发布标签，保证「最新版本」没有歧义。
var tagRe = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

type version struct {
	major, minor, patch int
}

func parseTag(tag string) (version, bool) {
	m := tagRe.FindStringSubmatch(tag)
	if m == nil {
		return version{}, false
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	return version{major, minor, patch}, true
}

func (v version) less(o version) bool {
	if v.major != o.major {
		return v.major < o.major
	}
	if v.minor != o.minor {
		return v.minor < o.minor
	}
	return v.patch < o.patch
}

// latestTag 从一组标签里挑出最大的正式版；按数值比较，v1.10.0 > v1.9.0。
func latestTag(tags []string) (string, bool) {
	var best string
	var bestV version
	found := false
	for _, t := range tags {
		v, ok := parseTag(t)
		if !ok {
			continue
		}
		if !found || bestV.less(v) {
			best, bestV, found = t, v, true
		}
	}
	return best, found
}
