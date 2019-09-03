# Config

Simple configuration module.  By default, will load a configuration from a file (optional, but must be JSON if 
provided), env vars, then flags.  A flag will be exposed for your app "-configFile" which lets the app declare where to 
find a config file.

## Usage

First, create a struct that is a composition of config.BaseConfiguration:

```go
type ExampleConfig struct {
	config.BaseConfiguration
	Text      string
}
```

Then, all config.Load on a pointer ref to an instance of the newly-created struct.  Ensure any
property in the instance is initialize to what you want the default value of that field to be:

```go
tc := TestConfig{
    Text: "defaultText",
}
cfg, err := config.Load(&tc)
if err != nil {
    log.Fatal(err)
}

tcResult := cfg.(*TestConfig)
```

In this example, it will load a config file from wherever the -configFile flag references.  If it is not
present, a config file will not be loaded.  If loaded, any value that is non-blank in the file will overwrite 
the matching value from the configuration instance initially passed.

Then, environment variables will be applied to the configuration.  For each property, the environment variable is
the upper-case name of the property, where dots are replaced with underscores.  So something like "OAuth.clientId" will 
be OAUTH_CLIENTID.

Finally, any flag (CLI argument) will be applied.  Each property is the dot property, where the dot is replaced by a 
dash.  So, something like "database.url" would be database-url as a flag.

### Using a Default Config File Path

To provide a default config file (to not require a flag/cli arg), you can initialize your config instance with
BaseConfiguration.ConfigFile:

```go
tc := TestConfig{
    BaseConfiguration: config.BaseConfiguration{
        ConfigFile: "./config.json",
    },
    Text: "",
}
```

### Environment Variable Prefixes

For environment variables, this module also supports an environment variable prefix.  To use this feature, provide
a second argument for config.Load (the uppercase value of what is provided will be used, with a trailing underscore):

```go
cfg, err := config.Load(&tc, "myapp")
```

This will make something like "MYAPP_TEXT" correspond to the "Text" property in the examples above.
