package main

import "os"

// You may want to read it from the conf
var DEFAULT_SERVER_PORT = "8080"
var SERVER_PORT = getServerPort()

// Read the port from the environment variable otherwise use the default value
func getServerPort() string {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = DEFAULT_SERVER_PORT
	}
	return port
}
