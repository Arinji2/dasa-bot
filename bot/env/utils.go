package env

import "strings"

func multiEnv(env string) []string {
	return strings.Split(env, ",")
}
