package common

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vechain/networkhub/network/node"
	nodegenesis "github.com/vechain/networkhub/network/node/genesis"
)

type fakeNode struct{ addr string }

func (f *fakeNode) Enode(string) (string, error)            { return "", nil }
func (f *fakeNode) SetExecArtifact(string)                  {}
func (f *fakeNode) GetConfigDir() string                    { return "" }
func (f *fakeNode) SetConfigDir(string)                     {}
func (f *fakeNode) GetDataDir() string                      { return "" }
func (f *fakeNode) SetDataDir(string)                       {}
func (f *fakeNode) SetID(string)                            {}
func (f *fakeNode) GetID() string                           { return "" }
func (f *fakeNode) GetExecArtifact() string                 { return "" }
func (f *fakeNode) GetKey() string                          { return "" }
func (f *fakeNode) GetGenesis() *nodegenesis.CustomGenesis  { return nil }
func (f *fakeNode) SetGenesis(*nodegenesis.CustomGenesis)   {}
func (f *fakeNode) SetAPIHost(string)                       {}
func (f *fakeNode) GetAPIHost() string                      { return "" }
func (f *fakeNode) GetAPIAddr() string                      { return f.addr }
func (f *fakeNode) SetAPIAddr(string)                       {}
func (f *fakeNode) GetAPICORS() string                      { return "*" }
func (f *fakeNode) GetP2PListenPort() int                   { return 0 }
func (f *fakeNode) SetP2PListenPort(int)                    {}
func (f *fakeNode) GetAdditionalArgs() map[string]string    { return nil }
func (f *fakeNode) SetAdditionalArgs(map[string]string)     {}
func (f *fakeNode) AddAdditionalArg(string, string)         {}
func (f *fakeNode) GetVerbosity() int                       { return 0 }
func (f *fakeNode) GetHTTPAddr() string                     { return f.addr }
func (f *fakeNode) GetFakeExecution() bool                  { return false }
func (f *fakeNode) HealthCheck(uint32, time.Duration) error { return nil }

func makeFakeNodes(n int) []node.Config {
	nodes := make([]node.Config, n)
	for i := 0; i < n; i++ {
		nodes[i] = &fakeNode{addr: fmt.Sprintf("http://127.0.0.1:%d", 8001+i)}
	}
	return nodes
}

func TestWaitForPeersConnection(t *testing.T) {
	prev := getPeerCount
	defer func() { getPeerCount = prev }()

	cases := []struct {
		name      string
		numNodes  int
		stubCount int
		stubErr   error
		timeout   time.Duration
		wantErr   bool
	}{
		{
			name:      "all peers connected quickly",
			numNodes:  3,
			stubCount: 2,
			stubErr:   nil,
			timeout:   time.Second,
			wantErr:   false,
		},
		{
			name:      "insufficient peers until timeout",
			numNodes:  2,
			stubCount: 0,
			stubErr:   nil,
			timeout:   150 * time.Millisecond,
			wantErr:   true,
		},
		{
			name:      "client error while fetching peers",
			numNodes:  2,
			stubCount: 0,
			stubErr:   errors.New("boom"),
			timeout:   150 * time.Millisecond,
			wantErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			getPeerCount = func(string) (int, error) { return tc.stubCount, tc.stubErr }
			nodes := makeFakeNodes(tc.numNodes)
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			err := WaitForPeersConnection(nodes, ctx)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
