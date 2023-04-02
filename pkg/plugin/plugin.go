package plugin

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-plugin"
	log "github.com/sirupsen/logrus"
	spireLog "github.com/spiffe/spire/pkg/common/log"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
)

var Clients []*plugin.Client

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"checks_provider": &ChecksProviderPlugin{},
}

func LoadPlugins() {
	lruLogger := &log.Logger{
		Out:       os.Stdout,
		Level:     log.DebugLevel,
		Formatter: &log.TextFormatter{},
	}
	logger := spireLog.NewHCLogAdapter(lruLogger, "plugin")

	homeDir, _ := os.UserHomeDir()
	pluginsDir := filepath.Join(homeDir, ".shipshape", "plugins")
	files, err := os.ReadDir(pluginsDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		Clients = append(Clients, plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig: handshakeConfig,
			Plugins:         pluginMap,
			Cmd:             exec.Command(file.Name()),
			Logger:          logger,
		}))
	}
}

func RegisterChecks() {
	for _, client := range Clients {
		// Connect via RPC
		rpcClient, err := client.Client()
		if err != nil {
			log.Fatal(err)
		}

		// Request the plugin
		raw, err := rpcClient.Dispense("checks_provider")
		if err != nil {
			log.Fatal(err)
		}

		checksProvider := raw.(ChecksProvider)
		providedChecks := checksProvider.Checks()
		for checkType, check := range providedChecks {
			config.ChecksRegistry[checkType] = check
		}
	}
}

func KillPlugins() {
	for _, client := range Clients {
		client.Kill()
	}
}
