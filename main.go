package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"gopkg.in/yaml.v2"
)

var dir string

type Resource interface {
	Process(resource []string, b []byte) []interface{}
	Conditions() []Condition
}

type Condition struct {
	ID       string `yaml:"id"`
	Alert    string `yaml:"alert,omitempty"`
	Warn     string `yaml:"warn,omitempty"`
	Duration int    `yaml:"duration"`
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s DIR", os.Args[0])
	}
	dir := os.Args[1]
	if _, err := os.Stat(path.Join(dir, "terraform.tfstate")); err != nil {
		log.Fatal(err)
	}

	err := os.Chdir(dir)
	if err != nil {
		log.Fatal(err)
	}

	resourceMap := getResources()
	//to prevent yaml returns []
	y := strings.TrimSpace(string(processResources(resourceMap)))
	if y == "[]" {
		fmt.Print("")
	} else {
		fmt.Print(y)
	}

}

func getResources() []string {
	cmd := exec.Command("bash", "-c", "terraform show | grep -E '^[a-zA-Z]' | tr -d ':'")
	b, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Split(string(b), "\n")
}

func processResources(resourceMap []string) []byte {
	b, err := ioutil.ReadFile(path.Join(dir, "terraform.tfstate"))
	if err != nil {
		log.Fatal(err)
	}

	var conditions []interface{}
	for _, resource := range resourceMap {
		if strings.Contains(resource, "aws_instance") {
			thing := Server{}
			conditions = append(conditions, thing.Process(resource, b)...)
		} else if strings.Contains(resource, "aws_sqs_queue") {
			thing := SQS{}
			conditions = append(conditions, thing.Process(resource, b)...)
		}
	}

	b2, err := yaml.Marshal(conditions)
	if err != nil {
		log.Fatal(err)
	}
	return b2
}
