package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type AppCondition struct {
	Details Condition `yaml:"app"`
}

type App struct{}

type DockerCompose struct {
	Version  string `yaml:"version"`
	Services map[string]struct {
		ContainerName string            `yaml:"container_name"`
		Image         string            `yaml:"image"`
		Labels        map[string]string `yaml:"labels"`
	} `yaml:"services"`
}

func getServices() []string {
	var services []string

	dc, err := ioutil.ReadFile(path.Join(dir, "docker-compose.yml"))
	if err != nil {
		return services
	}

	structure := DockerCompose{}
	err = yaml.Unmarshal(dc, &structure)
	if err != nil {
		return services
	}

	for key := range structure.Services {
		if structure.Services[key].Labels["alert"] == "manual" {
			continue
		}
		services = append(services, key)
	}

	return services
}

func (s App) Process(state string, resource string) (alerts []interface{}) {
	prefix := fmt.Sprintf("modules.#.resources.%v.", strings.Replace(resource, ".", "\\.", -1))
	name := queryJson(state, prefix+"primary.attributes.tags\\.Name")

	services := getServices()

	for _, service := range services {
		for _, v := range s.Conditions() {
			m := AppCondition{}
			m.Details.Alert = v.Alert
			m.Details.Warn = v.Warn
			m.Details.ID = name + "/" + service
			m.Details.Duration = v.Duration
			alerts = append(alerts, m)
		}
	}

	return alerts
}

func (s App) Conditions() []Condition {
	return []Condition{
		Condition{
			Alert:    "below 5 pulse",
			Duration: 30,
		},
	}
}