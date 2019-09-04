package config

import (
	"encoding/json"
	"flag"
	"github.com/markdicksonjr/dot"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
	"syscall"
)

// BaseConfiguration is to be composed into your custom configuration model
// so this module can pick up the "configFile" flag to tell us which config
// file should be loaded
type BaseConfiguration struct {
	ConfigFile string `json:"configFile"`
}

func Load(configWithDefaultValues interface{}, envPrefix ...string) (interface{}, error) {

	// first, grab the flags - they may contain info about files, etc, to grab other configs from
	flagCfg, err := copy(configWithDefaultValues)
	if err != nil {
		return nil, err
	}
	if err := applyFlags(flagCfg); err != nil {
		return nil, err
	}

	// grab the config file, defaulting to config.json
	configFile := dot.GetString(flagCfg, "configFile")

	// allow the config file to be optional (merge it into the default, if it's there)
	if len(configFile) > 0 {

		// read config.json
		file, err := ioutil.ReadFile(configFile)
		if err != nil {
			return configWithDefaultValues, err
		}
		var cfgFromFile interface{}

		if strings.HasSuffix(configFile, "json") {
			if err = json.Unmarshal(file, &cfgFromFile); err != nil {
				return configWithDefaultValues, err
			}
		} else {
			if err = yaml.Unmarshal(file, &cfgFromFile); err != nil {
				return configWithDefaultValues, err
			}
		}

		// allow a config file to overwrite the "default" config
		err = dot.Extend(configWithDefaultValues, cfgFromFile)
		if err != nil {
			return configWithDefaultValues, err
		}
	}

	// the order of precedence is file, env, flag

	// process env -> default/merged config + flags
	prefix := ""
	if len(envPrefix) > 0 {
		prefix = envPrefix[0]
	}
	if err := applyEnv(configWithDefaultValues, prefix); err != nil {
		return configWithDefaultValues, err
	}

	// process flags -> default/merged config
	err = dot.Extend(configWithDefaultValues, flagCfg)
	if err != nil {
		return configWithDefaultValues, err
	}

	return configWithDefaultValues, nil
}

func applyEnv(newCfg interface{}, prefix string) error {
	if prefix != "" {
		prefix = strings.ToUpper(prefix) + "_"
	}

	keys := dot.KeysRecursiveLeaves(newCfg)
	for _, key := range keys {
		k := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		if val, exist := syscall.Getenv(prefix + k); exist {
			if err := dot.Set(newCfg, key, val); err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO: ALLOW DESCS TO BE STATED SOMEHOW
func applyFlags(flagCfg interface{}) error {
	keys := dot.KeysRecursiveLeaves(flagCfg)
	for _, key := range keys {
		k := strings.ReplaceAll(key, ".", "-")

		dotVal, _ := dot.Get(flagCfg, k)
		isTypeSet := false
		if dotVal != nil {
			if dv, ok := dotVal.(bool); ok {
				isTypeSet = true
				flag.BoolVar(new(bool), k, dv, "")
			}
		}

		if !isTypeSet {
			val := dot.GetString(flagCfg, k)
			flag.StringVar(new(string), k, val, "")
		}
	}

	flag.Parse()

	var visitErr error
	flag.Visit(func(f *flag.Flag) {
		name := strings.Replace(f.Name, "-", ".", -1)

		if err := dot.Set(flagCfg, name, f.Value); err != nil {
			visitErr = err
		}
	})

	return visitErr
}

func copy(val interface{}) (interface{}, error) {
	marshalled, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}

	var copy interface{}
	if err := json.Unmarshal(marshalled, &copy); err != nil {
		return nil, err
	}

	return copy, nil
}
