package plugin

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/output"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "SHIPSHAPE_PLUGIN",
	MagicCookieValue: "28d1923e-f22f-41b7-9699-cdbf4a8e2faa",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"outputter": &output.OutputterPlugin{},
}

func Run() {
	// Create an hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins:         pluginMap,
		Cmd:             exec.Command("./build/plugin-sample"),
		Logger:          logger,
	})
	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Fatal(err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("outputter")
	if err != nil {
		log.Fatal(err)
	}

	outputter := raw.(output.Outputter)
	b, err := outputter.Output(&result.ResultList{
		Policies: map[string][]string{"policy1": {"rule1", "rule2"}},
		Results:  []result.Result{{Name: "result1", Passes: []string{"rule1"}}},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("output from plugin:", string(b))
}
