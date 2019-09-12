# Config

Simple configuration module.  By default, it will load a configuration from a file (optional, can be either JSON or YAML if 
provided), env vars, then flags.  A special flag will be exposed for your app "-configFile" which lets the app declare 
where to find a JSON or YAML config file.

## Usage

First, create a struct that is a composition of config.BaseConfiguration, along with your other configuration attributes:

```go
type ExampleConfig struct {
	config.BaseConfiguration
	Text      string
}
```

Then, call config.Load on a pointer ref to an instance of the newly-created struct.  Ensure any property in the instance 
is initialized to what you want the default value of that field to be.  Any slice should be initialize to an empty slice
for best results:

```go
tc := ExampleConfig{
    Text: "defaultText",
}
cfg, err := config.Load(&tc)
if err != nil {
    log.Fatal(err)
}

tcResult := cfg.(*ExampleConfig)
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
tc := ExampleConfig{
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
