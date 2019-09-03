package main

import (
	"github.com/markdicksonjr/config"
	"log"
)

type TestConfig struct {
	config.BaseConfiguration
	Text      string
	Debug     bool `json:"debug"`
	Primary   Classification
	Secondary *Classification `json:"Secondary,omitempty"`
}

type Classification struct {
	A string
	B string
}

// main will start the test app - to test, provide either a flag "-Primary-A BC" or a config file with Primary.A = "BC"
// this app asserts that the precedence of configs is what we expect
func main() {
	tc := TestConfig{
		BaseConfiguration: config.BaseConfiguration{
			//ConfigFile: "./config.json",
		},
		Text: "",
		Primary: Classification{
			A: "AC",
		},
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
