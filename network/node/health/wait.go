package health

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/thor/v2/thorclient"
)

var getPeerCount = func(httpAddr string) (int, error) {
	peers, err := thorclient.New(httpAddr).Peers()
	if err != nil {
		return 0, err
	}
	return len(peers), nil
}

// WaitForPeersConnection waits until every node sees all other nodes as peers.
func WaitForPeersConnection(nodes []node.Config, ctx context.Context) error {
	if len(nodes) == 0 {
		return nil
	}

	expected := len(nodes) - 1

	ctxWithTimeout := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctxWithTimeout, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
	}

	check := func() bool {
		for _, n := range nodes {
			count, err := getPeerCount(n.GetHTTPAddr())
			if err != nil || count < expected {
				return false
			}
		}
		return true
	}

	if check() {
		return nil
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctxWithTimeout.Done():
			return errors.New("timed out waiting for nodes to connect")
		case <-ticker.C:
			if check() {
				return nil
			}
		}
	}
}

func HealthCheck(block uint32, timeout time.Duration, httpAddr string) error {
	client := thorclient.New(httpAddr)
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			return fmt.Errorf("timeout waiting for node %s to be healthy", httpAddr)
		default:
			blk, err := client.Block(strconv.Itoa(int(block)))
			if err == nil && blk != nil {
				return nil
			}
			slog.Debug("waiting for node to be healthy", "node", httpAddr, "block", block, "error", err)
			time.Sleep(1 * time.Second)
		}
	}
}
