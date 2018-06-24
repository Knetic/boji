package main

import (
	"flag"
	"os"
	"fmt"
	"boji"
)

func main() {

	settings, err := parseFlags()
	if err != nil {
		fatal(err)
	}

	server := boji.NewServer(settings)
	server.Listen()
}

func parseFlags() (boji.ServerSettings, error) {

	var settings boji.ServerSettings

	flag.IntVar(&settings.Port, "p", 5170, "Port to serve on")
	flag.StringVar(&settings.Root, "r", "/var/lib/boji/data", "Path to root of served tree")
	settings.AdminUsername = coalesceEnv("BOJI_USER", "boji")
	settings.AdminPassword = coalesceEnv("BOJI_PASS", "boji")

	flag.Parse()

	return settings, nil
}

func coalesceEnv(env string, fallback string) string {

	value := os.Getenv(env)
	if value == "" {
		return fallback
	}
	return value
}

func fatal(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}