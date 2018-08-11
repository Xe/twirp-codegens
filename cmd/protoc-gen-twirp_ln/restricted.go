package main

import "strings"

func isRestricted(inp string) bool {
	for _, thing := range []string{"password", "token", "secret", "auth"} {
		if strings.Contains(inp, thing) {
			return true
		}
	}

	return false
}
