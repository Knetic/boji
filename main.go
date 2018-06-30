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

	err = os.MkdirAll(settings.Root, 0744)
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
	flag.StringVar(&settings.TLSCertPath, "c", "/etc/boji/certificate.crt", "Path to TLS certificate")
	flag.StringVar(&settings.TLSKeyPath, "k", "/etc/boji/server.key", "Path to TLS key file")
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