package main

import (
	"fmt"
	"strconv"
	"strings"
)

func isSecondVersionNewerThanFirstVersion(first, second string) (bool, error) {
	if prefix(first) != "" {
		if prefix(first) != prefix(second) {
			return false, fmt.Errorf("prefix mismatch: %s vs %s", prefix(first), prefix(second))
		}
	}
	if suffix(first) != "" {
		if suffix(first) != suffix(second) {
			return false, fmt.Errorf("suffix mismatch: %s vs %s", suffix(first), suffix(second))
		}
	}

	f := parseVersion(first)
	s := parseVersion(second)
	l := len(f)
	if len(s) > l {
		l = len(s)
	}
	for i := 0; i < l; i++ {
		fv := 0
		if i < len(f) {
			fv = f[i]
		}
		sv := 0
		if i < len(s) {
			sv = s[i]
		}
		if sv > fv {
			return true, nil
		}
		if sv < fv {
			return false, nil
		}
	}
	return false, nil
}

func parseVersion(v string) []int {
	if strings.HasPrefix(v, "v") {
		v = v[1:]
	}
	if i := strings.Index(v, "-"); i != -1 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			break
		}
		nums = append(nums, n)
	}
	return nums
}

func prefix(v string) string {
	if strings.HasPrefix(v, "v") {
		return "v"
	}
	return ""
}

func suffix(v string) string {
	if i := strings.Index(v, "-"); i != -1 {
		return v[i:]
	}
	return ""
}
