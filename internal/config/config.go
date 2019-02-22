package config

import (
	"encoding/json"
	"os"
	"strings"
)

var defConf = map[string]string{
	"PG_URL":            "postgres://postgres:password@localhost/postgres?sslmode=disable",
	"SERVER_ADDR":       ":8080",
	"DISABLE_PLAIN_LOG": "",
}

func get(key string) string {
	res := os.Getenv(key)
	if res == "" {
		res = defConf[key]
	}
	return res
}

func PrintAll() {
	r := make(map[string]string)
	for k := range defConf {
		r[k] = get(k)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(r)
}

func PG_URL() string {
	return get("PG_URL")
}

func SERVER_ADDR() string {
	return get("SERVER_ADDR")
}

func DISABLE_PLAIN_LOG() bool {
	x := get("DISABLE_PLAIN_LOG")
	return x != "" && strings.ToLower(x) != "false"
}
