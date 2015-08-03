package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const (
	configPath string = "./config.json"
)

var (
	// Config stores the global server
	// configuration parsed from the passed
	// config JSON file
	Config *Configuration

	// config holds the configuration
	// path
	config string
)

// Configuration holds configuration information
// for the service
//
// Hooks, used for the hooked analysis
// endpoint, are given as a map of id's
// to Hooks. (id's are strings that are
// specified within the passed JSON to the
// POST /task endpoint.) The default hook
// can be passed so you won't have to
// manually pass in a hookId when running
// a task. It defaults to the given hook
// when you only pass one hook in the config.
// When padding multiple, it defaults to
// a random hook.
type Configuration struct {
	Port       int16 `json:"port,omitempty"`
	portString string

	Hooks       map[string]Hook `json:"hooks,omitempty"`
	DefaultHook string          `json:"defaultHook,omitempty"`
}

// init grabs the config from the expected
// command line args. Else initializes with
// defaults
func init() {
	// holds config path
	flag.StringVar(&config, "conf", configPath, "Sets the server configuration filepath")
	flag.StringVar(&config, "C", configPath, "(shorthand for -conf)")
}

// ParseConfigFromFile must be run after
// the flag.Parse() function has been run
// to set the global configuration
// variable
func ParseConfigFromFile() error {
	path, err := filepath.Abs(config)
	if err != nil {
		return fmt.Errorf("ERROR: error generating absolute path from given config path. Does the file exist? %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("ERROR: error opening config file: %v", err)
	}

	// read file into buffer
	// 1KB should be enough for
	// a reasonable config file
	bytes := make([]byte, 1024)
	n, err := f.Read(bytes)
	if err != nil && err != io.EOF {
		return fmt.Errorf("ERROR: error reading config file into buffer: %v", err)
	}

	// unmarshal file into Config struct
	err = json.Unmarshal(bytes[:n], &Config)
	if err != nil {
		return fmt.Errorf("ERROR: error unmarshalling given config file into a Config struct: %v", err)
	}

	if Config.Port == 0 {
		Config.Port = 8080
	}

	Config.portString = fmt.Sprintf(":%v", Config.Port)

	if Config.DefaultHook == "" {
		for id := range Config.Hooks {
			Config.DefaultHook = id
		}
	}

	return nil
}

// ParseConfigFromURL pulls the config
// from a URL by making a get request
//
// This expects that the config path
// is, in fact, a URL routing to
// either HTTP or HTTPS
func ParseConfigFromURL() error {
	resp, err := http.Get(config)
	if err != nil {
		return fmt.Errorf("ERROR: error making GET request to given url: %v", config)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ERROR: error reading config data from GET url response. URL: %v", config)
	}

	err = json.Unmarshal(data, &Config)
	if err != nil {
		return fmt.Errorf("ERROR: error unmarshalling config data from response body. URL: %v", config)
	}

	if Config.Port == 0 {
		Config.Port = 8080
	}

	Config.portString = fmt.Sprintf(":%v", Config.Port)

	if Config.DefaultHook == "" {
		for id := range Config.Hooks {
			Config.DefaultHook = id
		}
	}

	return nil
}

// ParseConfig determines whether
// the config is at a URL or a file,
// unmarshals and munges the
// configuration, and returns
// any errors.
func ParseConfig() error {
	var err error

	// check if file
	u, err := url.Parse(config)
	if err != nil || !(u.Scheme == "http" || u.Scheme == "https") {
		err = ParseConfigFromFile()
	} else {
		err = ParseConfigFromURL()
	}

	if err != nil {
		return err
	}

	if Config.Port == 0 {
		Config.Port = 8080
	}

	Config.portString = fmt.Sprintf(":%v", Config.Port)

	if Config.DefaultHook == "" {
		for id := range Config.Hooks {
			Config.DefaultHook = id
		}
	}

	return nil
}
