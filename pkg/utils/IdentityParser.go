package utils

import "strings"

func IdentityParser(identity string) *map[string]string {
	r := make(map[string]string)
	if !strings.Contains(identity, ";") {
		v := strings.Split(identity, "=")
		r[v[0]] = v[1]
		return &r
	}

	for _, group := range strings.Split(identity, ";") {
		v := strings.Split(group, "=")
		r[v[0]] = v[1]
	}
	return &r
}
