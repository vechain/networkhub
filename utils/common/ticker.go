package common

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/vechain/thor/v2/thorclient"
)

type blockInfo struct {
	Number    uint32
	Timestamp uint64
}

type tickerClient interface {
	ExpandedBlockInfo(ref string) (*blockInfo, error)
}

type thorAdapter struct{ c *thorclient.Client }

func (a thorAdapter) ExpandedBlockInfo(ref string) (*blockInfo, error) {
	b, err := a.c.ExpandedBlock(ref)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	return &blockInfo{Number: b.Number, Timestamp: b.Timestamp}, nil
}

type Ticker struct {
	client tickerClient
}

func NewTicker(client *thorclient.Client) *Ticker {
	return &Ticker{client: thorAdapter{c: client}}
}

// NewTickerFrom allows injecting a custom client (useful for tests)
func NewTickerFrom(client tickerClient) *Ticker { return &Ticker{client: client} }

// Wait waits for the best block height to advance within the given timeout.
// It returns nil for the block to avoid coupling to external API types.
func (t *Ticker) Wait(timeout time.Duration) (any, error) {
	best, err := t.client.ExpandedBlockInfo("best")
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
			block, err := t.client.ExpandedBlockInfo("best")
			if err == nil && block != nil && best != nil && block.Number > best.Number {
				return nil, nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// WaitForBlock waits until the given block number is available (or an error/timeout occurs).
func (t *Ticker) WaitForBlock(blockNumber uint32) error {
	best, err := t.client.ExpandedBlockInfo("best")
	if err != nil {
		return err
	}
	if best != nil && blockNumber <= best.Number {
		return nil
	}
	if best != nil && best.Number == 0 { // edge case -> spinning up a new network with old genesis timestamps
		best.Timestamp = uint64(time.Now().Unix())
	}
	bestTs := uint64(time.Now().Unix())
	bestNum := uint32(0)
	if best != nil {
		bestTs = best.Timestamp
		bestNum = best.Number
	}
	expectedTime := bestTs + uint64(blockNumber-bestNum)*10
	timeout := time.Until(time.Unix(int64(expectedTime), 0).Add(20 * time.Second))
	tk := time.NewTicker(timeout)
	defer tk.Stop()

	slog.Info("waiting for block...", "block", blockNumber, "timeout", timeout.Seconds())

	for {
		select {
		case <-tk.C:
			return errors.New("timeout waiting for block")
		default:
			block, err := t.client.ExpandedBlockInfo(strconv.Itoa(int(blockNumber)))
			if err == nil && block != nil && block.Number >= blockNumber {
				return nil
			}
			time.Sleep(1 * time.Second)
			slog.Warn("waiting for block...", "block", blockNumber, "timeout", time.Until(time.Unix(int64(expectedTime), 0).Add(2*time.Second)))
		}
	}
}

type ConditionFunc func() (bool, error)

// WaitForCondition waits until conditionalFunc returns true or timeout occurs.
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
