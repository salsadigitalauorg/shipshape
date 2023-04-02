package plugin

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
)

type ChecksProvider interface {
	Checks() map[config.CheckType]func() config.Check
}

type ChecksProviderRPC struct{ client *rpc.Client }

func (g *ChecksProviderRPC) Checks() map[config.CheckType]func() config.Check {
	var resp map[config.CheckType]func() config.Check
	err := g.client.Call("Plugin.Checks", new(interface{}), &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

type ChecksProviderRPCServer struct {
	Impl ChecksProvider
}

func (s *ChecksProviderRPCServer) Checks(args interface{}, resp *map[config.CheckType]func() config.Check) error {
	*resp = s.Impl.Checks()
	return nil
}

type ChecksProviderPlugin struct {
	Impl ChecksProvider
}

func (p *ChecksProviderPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ChecksProviderRPCServer{Impl: p.Impl}, nil
}

func (ChecksProviderPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ChecksProviderRPC{client: c}, nil
}
