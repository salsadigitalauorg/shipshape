package docker_test

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	. "github.com/salsadigitalauorg/shipshape/pkg/docker"
)

func TestParse(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name               string
		fileStr            string
		envMap             map[string]string
		expectedBaseImages []BaseImage
		expectedError      string
	}{
		{
			name:               "empty",
			fileStr:            "",
			expectedBaseImages: []BaseImage(nil),
			expectedError:      "file with no instructions",
		},
		{
			name:    "simple",
			fileStr: "FROM php",
			expectedBaseImages: []BaseImage{
				{
					Image:         "php",
					ResolvedImage: "php",
					Tag:           "",
					ResolvedTag:   "latest",
				},
			},
			expectedError: "",
		},
		{
			name:    "simple/WithTag",
			fileStr: "FROM php:latest",
			expectedBaseImages: []BaseImage{
				{
					Image:         "php",
					ResolvedImage: "php",
					Tag:           "latest",
					ResolvedTag:   "latest",
				},
			},
			expectedError: "",
		},
		{
			name:    "simple/Multiple",
			fileStr: "FROM php:latest\nFROM alpine:3.12",
			expectedBaseImages: []BaseImage{
				{
					Image:         "php",
					ResolvedImage: "php",
					Tag:           "latest",
					ResolvedTag:   "latest",
				},
				{
					Image:         "alpine",
					ResolvedImage: "alpine",
					Tag:           "3.12",
					ResolvedTag:   "3.12",
				},
			},
			expectedError: "",
		},

		{
			name:    "singleArg",
			fileStr: "ARG PHP_VERSION=8.3\nFROM php:${PHP_VERSION}",
			expectedBaseImages: []BaseImage{
				{
					Image:         "php",
					ResolvedImage: "php",
					Tag:           "${PHP_VERSION}",
					ResolvedTag:   "8.3",
				},
			},
			expectedError: "",
		},
		{
			name:    "singleArg/EnvMapProvided",
			fileStr: "ARG PHP_VERSION=8.3\nFROM php:${PHP_VERSION}",
			envMap:  map[string]string{"PHP_VERSION": "8.4"},
			expectedBaseImages: []BaseImage{
				{
					Image:         "php",
					ResolvedImage: "php",
					Tag:           "${PHP_VERSION}",
					ResolvedTag:   "8.4",
				},
			},
			expectedError: "",
		},
		{
			name:    "singleArg/NoDefault",
			fileStr: "ARG PHP_VERSION\nFROM php:${PHP_VERSION}",
			expectedBaseImages: []BaseImage{
				{
					Image:         "php",
					ResolvedImage: "php",
					Tag:           "${PHP_VERSION}",
					ResolvedTag:   "latest",
				},
			},
			expectedError: "",
		},
		{
			name:    "singleArg/NoDefault/EnvMapProvided",
			fileStr: "ARG PHP_VERSION\nFROM php:${PHP_VERSION}",
			envMap:  map[string]string{"PHP_VERSION": "8.3"},
			expectedBaseImages: []BaseImage{
				{
					Image:         "php",
					ResolvedImage: "php",
					Tag:           "${PHP_VERSION}",
					ResolvedTag:   "8.3",
				},
			},
			expectedError: "",
		},
		{
			name:    "singleArg/NoBraces",
			fileStr: "ARG PHP_VERSION=8.3\nFROM php:$PHP_VERSION",
			expectedBaseImages: []BaseImage{
				{
					Image:         "php",
					ResolvedImage: "php",
					Tag:           "$PHP_VERSION",
					ResolvedTag:   "8.3",
				},
			},
			expectedError: "",
		},

		{
			name: "singleArg/ImageVar",
			fileStr: `ARG CLI_IMAGE
ARG PHP_IMAGE_VERSION=8.3

FROM ${CLI_IMAGE} as cli
FROM php:${PHP_IMAGE_VERSION}`,
			expectedBaseImages: []BaseImage{
				{
					Image:         "${CLI_IMAGE}",
					ResolvedImage: "",
					Tag:           "",
					ResolvedTag:   "latest",
				},
				{
					Image:         "php",
					ResolvedImage: "php",
					Tag:           "${PHP_IMAGE_VERSION}",
					ResolvedTag:   "8.3",
				},
			},
			expectedError: "",
		},
		{
			name: "singleArg/ImageVar/EnvMapProvided",
			fileStr: `ARG CLI_IMAGE
ARG PHP_IMAGE_VERSION=8.3

FROM ${CLI_IMAGE} as cli
FROM php:${PHP_IMAGE_VERSION}`,
			envMap: map[string]string{"CLI_IMAGE": "myproject-cli"},
			expectedBaseImages: []BaseImage{
				{
					Image:         "${CLI_IMAGE}",
					ResolvedImage: "myproject-cli",
					Tag:           "",
					ResolvedTag:   "latest",
				},
				{
					Image:         "php",
					ResolvedImage: "php",
					Tag:           "${PHP_IMAGE_VERSION}",
					ResolvedTag:   "8.3",
				},
			},
			expectedError: "",
		},

		{
			name: "updatedArgShouldNotImpactPrevious",
			fileStr: `ARG PHP_VERSION=8.3
FROM php:${PHP_VERSION}
ARG PHP_VERSION=8.4
FROM php:${PHP_VERSION}
`,
			expectedBaseImages: []BaseImage{
				{
					Image:         "php",
					ResolvedImage: "php",
					Tag:           "${PHP_VERSION}",
					ResolvedTag:   "8.3",
				},
				{
					Image:         "php",
					ResolvedImage: "php",
					Tag:           "${PHP_VERSION}",
					ResolvedTag:   "8.4",
				},
			},
			expectedError: "",
		},
	}

	currLogOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(currLogOut)
	logrus.SetOutput(io.Discard)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			baseImages, err := Parse([]byte(test.fileStr), test.envMap)
			if test.expectedError == "" {
				assert.Nil(err)
			} else {
				assert.Equal(test.expectedError, err.Error())
			}
			assert.Equal(test.expectedBaseImages, baseImages)
		})
	}

}
