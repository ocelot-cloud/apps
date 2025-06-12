package main

import (
	"strconv"
	"strings"
)

func isSecondVersionNewerThanFirstVersion(first, second string) bool {
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
			return true
		}
		if sv < fv {
			return false
		}
	}
	return false
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
