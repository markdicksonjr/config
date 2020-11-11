package config

import (
	"encoding/json"
	"flag"
	"github.com/markdicksonjr/dot"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"reflect"
	"strings"
	"syscall"
)

// BaseConfiguration is to be composed into your custom configuration model
// so this module can pick up the "configFile" flag to tell us which config
// file should be loaded
type BaseConfiguration struct {
	ConfigFile string `json:"configFile"`
}

// Load will apply changes to the configWithDefaultValues object provided by the user.  The
// precedence of value application (from most to least) is: flags, environment variables,
// file, then default configuration (provided by the user as the first parameter to this function).
// An optional "envPrefix" can be provided to this function to help differentiate its environment
// variable keys from those of the rest of the apps on the system.
func Load(configWithDefaultValues interface{}, envPrefix ...string) (interface{}, error) {

	// first, grab the flags - they may contain info about files, etc, to grab other configs from
	flagCfg, err := blankCopy(configWithDefaultValues)
	if err != nil {
		return nil, err
	}
	if err := applyFlags(flagCfg); err != nil {
		return nil, err
	}

	// grab the config file, defaulting to config.json
	configFile := dot.GetString(flagCfg, "configFile", "ConfigFile")

	// if the flags don't provide a config file, fall back to the user-provided default config
	if len(configFile) == 0 {
		configFile = dot.GetString(configWithDefaultValues, "configFile", "ConfigFile")
	}

	// allow the config file to be optional (merge it into the default, if it's there)
	if len(configFile) > 0 {

		// read config.json
		fileContents, err := ioutil.ReadFile(configFile)
		if err != nil {
			return configWithDefaultValues, err
		}

		cfgFromFile, err := blankCopy(configWithDefaultValues)
		if err != nil {
			return nil, err
		}

		if strings.HasSuffix(configFile, "json") {
			if err = json.Unmarshal(fileContents, &cfgFromFile); err != nil {
				return configWithDefaultValues, err
			}
		} else {
			if err = yaml.Unmarshal(fileContents, &cfgFromFile); err != nil {
				return configWithDefaultValues, err
			}
		}

		// allow a config fileContents to overwrite the "default" config
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

type boolVar struct {
	Ptr  *bool
	Key  string
	Name string
}

type stringVar struct {
	Ptr  *string
	Key  string
	Name string
}

type intVar struct {
	Ptr  *int
	Key  string
	Name string
}

// TODO: ALLOW DESCS TO BE STATED SOMEHOW
func applyFlags(flagCfg interface{}) error {
	keys := dot.KeysRecursiveLeaves(flagCfg)
	var boolFlags []boolVar
	var boolPtrFlags []boolVar
	var intFlags []intVar
	var intPtrFlags []intVar
	var stringFlags []stringVar
	var stringPtrFlags []stringVar

	for _, key := range keys {
		dotVal, _ := dot.Get(flagCfg, key)
		if dotVal == nil {
			continue
		}

		keys := strings.Split(key, ".")
		for i, k := range keys {
			keys[i] = strings.ToLower(string(k[0])) + k[1:]
		}
		camelKey := strings.Join(keys, "-")

		if ds, ok := dotVal.(string); ok {
			strFlag := stringVar{
				Name: camelKey,
				Key:  key,
				Ptr:  nil,
			}
			strFlag.Ptr = flag.String(strFlag.Name, ds, "")
			stringFlags = append(stringFlags, strFlag)
			continue
		}

		if dv, ok := dotVal.(bool); ok {
			boolFlag := boolVar{
				Name: camelKey,
				Key:  key,
				Ptr:  nil,
			}
			boolFlag.Ptr = flag.Bool(boolFlag.Name, dv, "")
			boolFlags = append(boolFlags, boolFlag)
			continue
		}

		if _, ok := dotVal.(*bool); ok {
			empty := false

			boolFlag := boolVar{
				Name: camelKey,
				Key:  key,
				Ptr:  &empty,
			}
			flag.BoolVar(&empty, boolFlag.Name, false, "")
			boolPtrFlags = append(boolPtrFlags, boolFlag)
			continue
		}

		if dv, ok := dotVal.(int); ok {
			intFlag := intVar{
				Name: camelKey,
				Key:  key,
				Ptr:  nil,
			}
			intFlag.Ptr = flag.Int(intFlag.Name, dv, "")
			intFlags = append(intFlags, intFlag)
			continue
		}

		if _, ok := dotVal.(*int); ok {
			empty := 0

			intFlag := intVar{
				Name: camelKey,
				Key:  key,
				Ptr:  &empty,
			}
			flag.IntVar(&empty, intFlag.Name, 0, "")
			intPtrFlags = append(intPtrFlags, intFlag)
			continue
		}

		if _, ok := dotVal.(*string); ok {
			empty := ""

			strFlag := stringVar{
				Name: camelKey,
				Key:  key,
				Ptr:  &empty,
			}
			flag.StringVar(&empty, strFlag.Name, "", "")
			stringPtrFlags = append(stringPtrFlags, strFlag)
			continue
		}

		// TODO: more types
	}

	flag.Parse()

	// loop through booleans and apply them
	for _, boolKey := range boolFlags {
		if err := dot.Set(flagCfg, boolKey.Key, *boolKey.Ptr); err != nil {
			return err
		}
	}

	for _, boolPtrKey := range boolPtrFlags {
		if err := dot.Set(flagCfg, boolPtrKey.Key, boolPtrKey.Ptr); err != nil {
			return err
		}
	}

	// loop through ints and apply them
	for _, intKey := range intFlags {
		if err := dot.Set(flagCfg, intKey.Key, *intKey.Ptr); err != nil {
			return err
		}
	}

	for _, intPtrKey := range intPtrFlags {
		if err := dot.Set(flagCfg, intPtrKey.Key, intPtrKey.Ptr); err != nil {
			return err
		}
	}

	// loop through strings and apply them
	for _, stringKey := range stringFlags {
		if err := dot.Set(flagCfg, stringKey.Key, *stringKey.Ptr); err != nil {
			return err
		}
	}

	for _, strPtrKey := range stringPtrFlags {
		if err := dot.Set(flagCfg, strPtrKey.Key, strPtrKey.Ptr); err != nil {
			return err
		}
	}

	return nil
}

func blankCopy(val interface{}) (interface{}, error) {
	_, canBeMap := val.(map[string]interface{})
	if canBeMap {
		finalMap := make(map[string]interface{})
		recursiveLeavesKeys := dot.KeysRecursiveLeaves(val)
		for _, key := range recursiveLeavesKeys {
			dotVal, _ := dot.Get(val, key)
			if dotVal == nil {
				continue
			}
			dotType := reflect.TypeOf(dotVal)

			if dotType.Kind() == reflect.Ptr {

				// save the zero-d version of the key to the outbound map
				asZero := reflect.Zero(dotType.Elem())
				if err := dot.Set(finalMap, key, asZero.Interface()); err != nil {
					return nil, err
				}
				continue
			}

			// save the zero-d version of the key to the outbound map
			asZero := reflect.Zero(dotType)
			if err := dot.Set(finalMap, key, asZero.Interface()); err != nil {
				return nil, err
			}
		}

		return finalMap, nil
	}

	return reflect.New(reflect.TypeOf(val).Elem()).Elem().Addr().Interface(), nil
}
