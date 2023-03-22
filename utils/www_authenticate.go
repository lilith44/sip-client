package utils

import "strings"

func ParseWWWAuthenticate(auth string) map[string]string {
	m := make(map[string]string)
	if auth == "" {
		return m
	}

	ss := strings.Split(auth, ",")
	for i := range ss {
		kv := strings.Split(ss[i], "=")

		if i == 0 {
			m[kv[0][strings.LastIndex(kv[0], " ")+1:]] = strings.Trim(kv[1], `"`)
		} else {
			m[strings.TrimLeft(kv[0], " ")] = strings.Trim(kv[1], `"`)
		}
	}
	return m
}
