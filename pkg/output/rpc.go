package output

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

// OutputterRPC is the client side implementation of the interface which is used
// by the host to send data to the plugin and receive the result.
type OutputterRPC struct{ client *rpc.Client }

func (or *OutputterRPC) Output(rl *result.ResultList) ([]byte, error) {
	var resp []byte
	err := or.client.Call("Plugin.Output", rl, &resp)
	return resp, err
}

// OutputterRPCServer is the server representation of our interface which is
// used by the plugin to receive data from the host and return the result.
type OutputterRPCServer struct {
	Impl Outputter
}

func (os *OutputterRPCServer) Output(rl *result.ResultList, resp *[]byte) error {
	b, err := os.Impl.Output(rl)
	*resp = b
	return err
}

// OutputterPlugin is what exposes the Outputter interface as a plugin.
type OutputterPlugin struct {
	Impl Outputter
}

func (p *OutputterPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &OutputterRPCServer{Impl: p.Impl}, nil
}

func (OutputterPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &OutputterRPC{client: c}, nil
}
