package common

import (
	"errors"
	"fmt"

	"log/slog"
	"strconv"
	"time"

	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/thorclient"
	"github.com/vechain/thor/v2/thorclient/httpclient"
)

type Ticker struct {
	client *thorclient.Client
}

func NewTicker(client *thorclient.Client) *Ticker {
	return &Ticker{
		client: client,
	}
}

// Wait waits for a new best block to be available
func (t *Ticker) Wait(timeout time.Duration) (*api.JSONExpandedBlock, error) {
	best, err := t.client.ExpandedBlock("best")
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			return nil, errors.New("timeout waiting for block")
		default:
			block, err := t.client.ExpandedBlock("best")
			if err == nil && block != nil && block.Number > best.Number {
				return block, nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (t *Ticker) WaitForBlock(blockNumber uint32) error {
	best, err := t.client.ExpandedBlock("best")
	if err != nil {
		return err
	}
	if blockNumber <= best.Number {
		return nil
	}
	if best.Number == 0 { // edge case -> spinning up a new network with old genesis timestamps
		best.Timestamp = uint64(time.Now().Unix())
	}
	expectedTime := best.Timestamp + uint64(blockNumber-best.Number)*10
	timeout := time.Until(time.Unix(int64(expectedTime), 0).Add(20 * time.Second))
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	slog.Info("waiting for block...", "block", blockNumber, "timeout", timeout.Seconds())

	for {
		select {
		case <-ticker.C:
			return errors.New("timeout waiting for block")
		default:
			block, err := t.client.ExpandedBlock(strconv.Itoa(int(blockNumber)))
			if block != nil && block.Number >= blockNumber {
				return nil
			}
			if err != nil && !errors.Is(err, httpclient.ErrNotFound) {
				return fmt.Errorf("unexpected error getting block: %w", err)
			}
			time.Sleep(1 * time.Second)
			slog.Warn("waiting for block...", "block", blockNumber, "timeout", time.Until(time.Unix(int64(expectedTime), 0).Add(2*time.Second)))
		}
	}
}

type ConditionFunc func() (bool, error)

func (t *Ticker) WaitForCondition(timeout time.Duration, conditionalFunc ConditionFunc) error {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			return errors.New("timeout waiting for condition")
		default:
			resp, err := conditionalFunc()
			if err != nil {
				return err
			}
			if resp {
				return nil
			}
			time.Sleep(10 * time.Second)
		}
	}
}
