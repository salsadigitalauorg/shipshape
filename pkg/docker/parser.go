package docker

import (
	"bytes"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/env"
)

type arg struct {
	key         string
	defVal      string
	resolvedVal string
}

type BaseImage struct {
	Image         string `json:"image"`
	ResolvedImage string `json:"resolvedImage"`

	Tag         string `json:"tag"`
	ResolvedTag string `json:"resolvedTag"`

	args map[string]string
}

func (a *arg) resolve(envMap map[string]string) {
	if val, ok := envMap[a.key]; ok {
		a.resolvedVal = val
	} else {
		a.resolvedVal = a.defVal
	}
}

func (b *BaseImage) resolve() error {
	resI, err := env.ResolveValue(b.args, b.Image)
	if err != nil {
		log.WithError(err).Error("could not resolve image name")
		return err
	}
	b.ResolvedImage = resI

	resT, err := env.ResolveValue(b.args, b.Tag)
	if err != nil {
		log.WithError(err).Error("could not resolve image tag")
		return err
	}
	b.ResolvedTag = resT
	if b.ResolvedTag == "" {
		b.ResolvedTag = "latest"
	}

	b.args = nil
	return nil
}

func (b BaseImage) String() string {
	return b.ResolvedImage + ":" + b.ResolvedTag
}

func Parse(file []byte, envMap map[string]string) ([]BaseImage, error) {
	dockerfile, err := parser.Parse(bytes.NewBuffer(file))
	if err != nil {
		log.WithError(err).Error("could not parse Dockerfile")
		return nil, err
	}

	var argsMap map[string]string
	baseImages := []BaseImage{}
	for _, child := range dockerfile.AST.Children {
		if strings.ToLower(child.Value) == "arg" {
			if argsMap == nil {
				argsMap = map[string]string{}
			}

			rawVal := child.Next.Value
			key := rawVal
			defVal := ""
			if strings.Contains(rawVal, "=") {
				key = strings.Split(rawVal, "=")[0]
				defVal = strings.Split(rawVal, "=")[1]
			}
			a := arg{key: key, defVal: defVal}
			a.resolve(envMap)
			argsMap[key] = a.resolvedVal
		}

		if strings.ToLower(child.Value) == "from" {
			rawVal := child.Next.Value
			image := rawVal
			tag := ""
			if strings.Contains(rawVal, ":") {
				image = strings.Split(rawVal, ":")[0]
				tag = strings.Split(rawVal, ":")[1]
			}
			i := BaseImage{Image: image, Tag: tag, args: argsMap}
			err := i.resolve()
			if err != nil {
				return nil, err
			}
			baseImages = append(baseImages, i)
		}
	}
	return baseImages, nil
}
