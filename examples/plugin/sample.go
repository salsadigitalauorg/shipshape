package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/output"
	shipshape_plugin "github.com/salsadigitalauorg/shipshape/pkg/plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

type SampleOutput struct {
	logger hclog.Logger
}

func (o *SampleOutput) Output(rl *result.ResultList) ([]byte, error) {
	o.logger.Debug("message from SampleOutput.Output")
	b := bytes.Buffer{}
	fmt.Fprintln(&b, "ResultList:", *rl)
	return b.Bytes(), nil
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	outputter := &SampleOutput{
		logger: logger,
	}
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		"outputter": &output.OutputterPlugin{Impl: outputter},
	}

	logger.Debug("message from plugin", "foo", "bar")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shipshape_plugin.Handshake,
		Plugins:         pluginMap,
	})
}
