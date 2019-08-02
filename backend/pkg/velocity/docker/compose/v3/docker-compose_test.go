package v3_test

import (
	"reflect"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	v3 "github.com/velocity-ci/velocity/backend/pkg/velocity/docker/compose/v3"
)

func TestDockerComposeYamlUnmarshal(t *testing.T) {
	dockerComposeYaml := `
---
version: '3'

services:

  a:
    image: busybox
    command: ["ping", "-c", "3", "c"]
    volumes:
      - "./:/app"
    links:
      - "b"
    environment:
      HELLO: WORLD
      num: 1
      bool: true
      float: !!float 1.234

  b:
    image: alpine
    command: /bin/sh -c 'sleep 10s'
    networks:
      default:
        aliases:
          - "c"
    environment:
      - HELLO=WORLD
  
  c:
    build: .
  
  d:
    build:
      context: "app"
      dockerfile: "app.Dockerfile"
`
	var dockerComposeConf v3.DockerComposeYaml
	err := yaml.Unmarshal([]byte(dockerComposeYaml), &dockerComposeConf)
	assert.Nil(t, err)

	expectedDockerComposeConf := v3.DockerComposeYaml{
		Version: "3",
		Services: map[string]v3.DockerComposeService{
			"a": {
				Image:   "busybox",
				Command: v3.DockerComposeServiceCommand{"ping", "-c", "3", "c"},
				Volumes: []string{"./:/app"},
				Links:   []string{"b"},
				Environment: v3.DockerComposeServiceEnvironment{
					"HELLO": "WORLD",
					"num":   "1",
					"bool":  "true",
					"float": "1.234",
				},
			},
			"b": {
				Image:   "alpine",
				Command: v3.DockerComposeServiceCommand{"/bin/sh", "-c", "sleep 10s"},
				Networks: map[string]v3.DockerComposeServiceNetwork{
					"default": {
						Aliases: []string{"c"},
					},
				},
				Environment: v3.DockerComposeServiceEnvironment{
					"HELLO": "WORLD",
				},
			},
			"c": {
				Build: v3.DockerComposeServiceBuild{
					Context:    ".",
					Dockerfile: "Dockerfile",
				},
			},
			"d": {
				Build: v3.DockerComposeServiceBuild{
					Context:    "app",
					Dockerfile: "app.Dockerfile",
				},
			},
		},
	}

	assert.Equal(t, expectedDockerComposeConf, dockerComposeConf)
}

func TestGetServiceOrder(t *testing.T) {
	type args struct {
		services     map[string]v3.DockerComposeService
		serviceOrder []string
	}

	services := map[string]v3.DockerComposeService{
		"proxy": {
			Links: []string{
				"backend",
				"frontend",
			},
		},
		"database": {
			Links: []string{
				"redis",
			},
		},
		"redis": {
			Links: []string{
				"frontend",
			},
		},
		"frontend": {
			Links: []string{},
		},
		"backend": {
			Links: []string{"database"},
		},
	}

	services2 := map[string]v3.DockerComposeService{
		"database": {
			Links: []string{
				"redis",
			},
		},
		"proxy": {
			Links: []string{
				"backend",
				"frontend",
			},
		},
		"redis": {
			Links: []string{
				"frontend",
			},
		},
		"frontend": {
			Links: []string{},
		},
		"backend": {
			Links: []string{"database"},
		},
	}

	services3 := map[string]v3.DockerComposeService{
		"redis": {
			Links: []string{
				"frontend",
			},
		},
		"database": {
			Links: []string{
				"redis",
			},
		},
		"frontend": {
			Links: []string{},
		},
		"backend": {
			Links: []string{"database"},
		},
		"proxy": {
			Links: []string{
				"backend",
				"frontend",
			},
		},
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "1",
			args: args{
				services:     services,
				serviceOrder: []string{},
			},

			want: []string{"frontend", "redis", "database", "backend", "proxy"},
		},
		{
			name: "2",
			args: args{
				services:     services2,
				serviceOrder: []string{},
			},

			want: []string{"frontend", "redis", "database", "backend", "proxy"},
		},
		{
			name: "3",
			args: args{
				services:     services3,
				serviceOrder: []string{},
			},

			want: []string{"frontend", "redis", "database", "backend", "proxy"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v3.GetServiceOrder(tt.args.services, tt.args.serviceOrder); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getServiceOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}
