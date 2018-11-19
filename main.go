package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

var dir string

type Resource interface {
	Process(resource []string, b []byte) []interface{}
	Conditions() []Condition
}

// Condition alert conditions
type Condition struct {
	ID       string `yaml:"id"`
	Alert    string `yaml:"alert,omitempty"`
	Warn     string `yaml:"warn,omitempty"`
	Duration int    `yaml:"duration"`
}

func main() {

	if len(os.Args) != 2 {
		log.Println("Version: v2.2")
		log.Fatalf("Usage: %s DIR", os.Args[0])
	}

	dir = os.Args[1]
	if _, err := os.Stat(path.Join(dir, "terraform.tfstate")); err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadFile(path.Join(dir, "terraform.tfstate"))
	if err != nil {
		log.Fatal(err)
	}

	resources := getResources(string(b))
	fmt.Print(string(processResources(resources)))
}

func getResources(state string) (resources []string) {
	if result := gjson.Get(state, "modules.#.resources").Array(); len(result) > 0 {
		for _, v := range result {
			for k, i := range v.Map() {
				// ignore all data resource
				if !strings.HasPrefix(k, "data.") {
					resources = append(resources, i.Raw)
				}
			}
		}
	}
	sort.Strings(resources)
	return resources
}

func processResources(resources []string) (b2 []byte) {
	var conditions []interface{}
	for _, resource := range resources {
		if gjson.Get(resource, "type").String() == "aws_instance" {

			fmt.Println("There is a aws_instance")
			server := Server{}
			conditions = append(conditions, server.Process(resource)...)

			app := App{}
			conditions = append(conditions, app.Process(resource)...)

		} else if gjson.Get(resource, "type").String() == "aws_sqs_queue" {
			sqs := SQS{}
			conditions = append(conditions, sqs.Process(resource)...)
		}
	}

	if len(conditions) > 0 {
		b2, err := yaml.Marshal(conditions)
		if err != nil {
			log.Fatal(err)
		} else {
			return b2
		}
	}
	return b2
}
