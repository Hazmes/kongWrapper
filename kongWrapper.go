// Package kongWrapper wraps alecthomas/kong to provide some utility around it.
// It allows to read in config through either command line or a config file.
// The config file can be provided via command line or env variable.
// Hieararchy: CLI <-- configPathCLI <-- configPathENV
package kongWrapper

import (
	"fmt"
	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
	"os"
	"path"
)

// HasConfig needs to implement a method called GetConfigPath
// GetConfigPath should return the field in the CLI definition representing a path to a config
type HasConfig interface {
	GetConfigPath() string
}

// Parse takes a struct type that reflects the config structure you want to load.
// The second argument is an env variable with the path to a config file.
// The struct type has to implement the HasConfig interface.
func Parse(cli HasConfig, configEnv string) error {

	// create new parser for the cli struct
	parser, err := kong.New(cli)
	if err != nil {
		return err
	}

	// parse cli arguments
	_, err = parser.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	// get config path from cli
	configPath := cli.GetConfigPath()
	if configPath == "" {
		// if config path not provided via cli, try to get it via env var
		if envPath, ok := configFileFromENV(configEnv); ok {
			configPath = envPath
		}
	}

	// if config path is present, create parser based on filetype
	if configPath != "" {
		parser, err = fromFileEnding(cli, configPath)
		if err != nil {
			return err
		}
	}

	// reparse cli arguments to overwrite config file
	_, err = parser.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	return nil
}

// fromFileEnding creates a new kong parser based on file ending
func fromFileEnding(cli HasConfig, configPath string) (*kong.Kong, error) {
	switch path.Ext(configPath) {
	case ".json":
		return kong.New(cli, kong.Configuration(kong.JSON, configPath))
	case ".yaml", ".yml":
		return kong.New(cli, kong.Configuration(kongyaml.Loader, configPath))
	default:
		return nil, fmt.Errorf("unsupported config format: %s", configPath)
	}
}

// configFileFromENV looks for an env variable containing a path to a config file
func configFileFromENV(configEnv string) (string, bool) {
	val, ok := os.LookupEnv(configEnv)

	if ok && val != "" {
		return val, true
	}

	return "", false
}
