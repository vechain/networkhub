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

// Wait waits for a new best block to be available within the given timeout.
func (t *Ticker) Wait(timeout time.Duration) (any, error) {
	best, err := t.client.ExpandedBlockInfo("best")
	if err != nil {
		return nil, err
	}
	err = t.WaitForCondition(timeout, func() (bool, error) {
		b, e := t.client.ExpandedBlockInfo("best")
		if e != nil || b == nil || best == nil {
			return false, nil
		}
		return b.Number > best.Number, nil
	})
	return nil, err
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
	expectedTime := time.Unix(int64(bestTs), 0).Add(time.Duration(blockNumber-bestNum) * 10 * time.Second)
	timeout := min(max(time.Until(expectedTime), 300*time.Millisecond), 2*time.Second)

	slog.Info("waiting for block...", "block", blockNumber, "deadline", timeout.String())

	return t.WaitForCondition(timeout, func() (bool, error) {
		b, e := t.client.ExpandedBlockInfo(strconv.Itoa(int(blockNumber)))
		if e != nil || b == nil {
			return false, nil
		}
		return b.Number >= blockNumber, nil
	})
}

type ConditionFunc func() (bool, error)

// WaitForCondition waits until conditionalFunc returns true or timeout occurs.
func (t *Ticker) WaitForCondition(timeout time.Duration, conditionalFunc ConditionFunc) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	for {
		resultCh := make(chan struct {
			ok  bool
			err error
		}, 1)

		go func() {
			ok, err := conditionalFunc()
			resultCh <- struct {
				ok  bool
				err error
			}{ok: ok, err: err}
		}()

		select {
		case <-deadline.C:
			return errors.New("timeout waiting for condition")
		case res := <-resultCh:
			if res.err != nil {
				return res.err
			}
			if res.ok {
				return nil
			}
			// brief pause before next attempt to avoid hot loop
			time.Sleep(100 * time.Millisecond)
		}
	}
}
