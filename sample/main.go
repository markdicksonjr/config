package main

import (
	"github.com/markdicksonjr/config"
	"log"
)

// TestConfig reflects an example of how an app's configuration should look:
// it is composed of config.BaseConfiguration, as well as additional attributes
type TestConfig struct {
	config.BaseConfiguration
	Text      string
	Debug     bool    `json:"debug"`
	Version   *bool   `json:"version,omitempty"`
	Name      *string `json:"name,omitempty"`
	Primary   Classification
	Secondary *Classification `json:"Secondary,omitempty"`
}

// Classification is a simple struct within TestConfig.  It's a simple 2-tuple
type Classification struct {
	A string
	B string
}

// main will start the test app - to test, provide either a flag "-Primary-A BC" or a config file with Primary.A = "BC"
// this app asserts that the precedence of configs is what we expect
func main() {
	isTrue := true
	tc := TestConfig{
		BaseConfiguration: config.BaseConfiguration{
			//ConfigFile: "./config.json",
		},
		Text: "",
		Primary: Classification{
			A: "AC",
		},
		Name: nil,
		Version: &isTrue,
	}
	fg, err := config.Load(&tc)
	if err != nil {
		log.Fatal(err)
	}

	tcResult := fg.(*TestConfig)
	if tcResult.Primary.A != "BC" {
		log.Fatal("Primary.A was not 'BC'")
	}
	log.Printf("Completed successfully")
}
