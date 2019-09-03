package config

import (
	"encoding/json"
	"flag"
	"github.com/markdicksonjr/dot"
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

		if err = json.Unmarshal(file, &cfgFromFile); err != nil {
			return configWithDefaultValues, err
		}

		// allow a config file to overwrite the "default" config
		mergedCfg, err := dot.Extend(configWithDefaultValues, cfgFromFile)
		if err != nil {
			return configWithDefaultValues, err
		}

		configWithDefaultValues = mergedCfg
	}

	// the order of precedence is file, env, flag

	// process flags -> default/merged config
	newCfg, err := dot.Extend(configWithDefaultValues, flagCfg)
	if err != nil {
		return configWithDefaultValues, err
	}

	// process env -> default/merged config + flags
	prefix := ""
	if len(envPrefix) > 0 {
		prefix = envPrefix[0]
	}
	if err := applyEnv(newCfg, prefix); err != nil {
		return configWithDefaultValues, err
	}

	cfgStr, err := json.Marshal(newCfg)
	if err != nil {
		return configWithDefaultValues, err

	}

	if err := json.Unmarshal(cfgStr, &configWithDefaultValues); err != nil {
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
	hash := map[string]*string{}
	for _, key := range keys {
		k := strings.ReplaceAll(key, ".", "-")
		hash[k] = new(string)
		val := dot.GetString(flagCfg, k)
		flag.StringVar(hash[k], k, val, "")
	}

	flag.Parse()

	var visitErr error
	flag.Visit(func(f *flag.Flag) {
		name := strings.Replace(f.Name, "-", ".", -1)
		if err := dot.Set(flagCfg, name, f.Value.String()); err != nil {
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
