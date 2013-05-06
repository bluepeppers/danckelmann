package config

import (
	"log"
	"fmt"
	"os"

	"github.com/bluepeppers/allegro"
)

func GetString(conf *allegro.Config, sec, key string, def string) string {
	val, ok := conf.Get(sec, key)
	if !ok {
		return def
	}
	return val
}

func GetInt(conf *allegro.Config, sec, key string, def int) int {
	var val int
	vals, ok := conf.Get(sec, key)
	// Don't want to log something for every missing option
	if !ok {
		return def
	}
	_, err := fmt.Sscanf(vals, "%d", &val)
	if err != nil {
		log.Printf("%s.%s=%q not parseable as integer", sec, key, vals)
		log.Printf("Defaulting to %s.%s=%d", sec, key, def)
		return def
	}
	return val
}

func GetBool(conf *allegro.Config, sec, key string, def bool) bool {
	var val bool
	vals, ok := conf.Get(sec, key)
	if !ok {
		return def
	}
	_, err := fmt.Sscanf(vals, "%t", &val)
	if err != nil {
		log.Printf("%s.%s=%q not parseable as boolean", sec, key, vals)
		log.Printf("Defaulting to %s.%s=%t", sec, key, def)
		return def
	}
	return val
}

func exists(fname string) bool {
	_, err := os.Open(fname)
	return !os.IsNotExist(err)
}

func LoadUserConfig(configLocation string) *allegro.Config {
	fname := os.ExpandEnv(configLocation)
	if exists(fname) {
		return allegro.LoadConfig(fname)
	}
	return allegro.CreateConfig()
}
