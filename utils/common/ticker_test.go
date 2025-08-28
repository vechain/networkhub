package common

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeTickerClient struct {
	blocks   map[string]*blockInfo
	calls    int
	errByRef map[string]error
}

func (f *fakeTickerClient) ExpandedBlockInfo(ref string) (*blockInfo, error) {
	f.calls++
	if err, ok := f.errByRef[ref]; ok && err != nil {
		return nil, err
	}
	if b, ok := f.blocks[ref]; ok {
		return b, nil
	}
	return nil, errors.New("not found")
}

func TestTicker_Wait(t *testing.T) {
	cases := []struct {
		name    string
		setup   func() *Ticker
		timeout time.Duration
		wantErr bool
	}{
		{
			name: "advances within timeout",
			setup: func() *Ticker {
				f := &fakeTickerClient{blocks: map[string]*blockInfo{"best": {Number: 1}}}
				tk := &Ticker{client: f}
				// advance block after 100ms
				go func() { time.Sleep(100 * time.Millisecond); f.blocks["best"] = &blockInfo{Number: 2} }()
				return tk
			},
			timeout: 500 * time.Millisecond,
			wantErr: false,
		},
		{
			name: "does not advance before timeout",
			setup: func() *Ticker {
				return &Ticker{client: &fakeTickerClient{blocks: map[string]*blockInfo{"best": {Number: 5}}}}
			},
			timeout: 150 * time.Millisecond,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tk := tc.setup()
			_, err := tk.Wait(tc.timeout)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTicker_WaitForBlock(t *testing.T) {
	notFound := errors.New("not found")
	cases := []struct {
		name    string
		setup   func() *Ticker
		block   uint32
		wantErr bool
	}{
		{
			name: "already at or past block",
			setup: func() *Ticker {
				return &Ticker{client: &fakeTickerClient{blocks: map[string]*blockInfo{"best": {Number: 10}}}}
			},
			block:   9,
			wantErr: false,
		},
		{
			name: "reaches block before timeout",
			setup: func() *Ticker {
				f := &fakeTickerClient{blocks: map[string]*blockInfo{"best": {Number: 1, Timestamp: uint64(time.Now().Unix())}}, errByRef: map[string]error{"2": notFound}}
				tk := &Ticker{client: f}
				go func() {
					time.Sleep(200 * time.Millisecond)
					f.blocks["2"] = &blockInfo{Number: 2}
					delete(f.errByRef, "2") // block now available
				}()
				return tk
			},
			block:   2,
			wantErr: false,
		},
		{
			name: "timeout before block is available",
			setup: func() *Ticker {
				return &Ticker{client: &fakeTickerClient{blocks: map[string]*blockInfo{"best": {Number: 1, Timestamp: uint64(time.Now().Unix())}}, errByRef: map[string]error{"2": notFound}}}
			},
			block:   2,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tk := tc.setup()
			// shrink the internal timeout window by faking smaller time delta via Timestamp
			err := tk.WaitForBlock(tc.block)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTicker_WaitForCondition(t *testing.T) {
	cases := []struct {
		name    string
		timeout time.Duration
		cond    func() (bool, error)
		wantErr bool
	}{
		{
			name:    "condition satisfied",
			timeout: 500 * time.Millisecond,
			cond:    func() (bool, error) { return true, nil },
			wantErr: false,
		},
		{
			name:    "condition errors",
			timeout: 500 * time.Millisecond,
			cond:    func() (bool, error) { return false, errors.New("boom") },
			wantErr: true,
		},
		{
			name:    "timeout before condition",
			timeout: 150 * time.Millisecond,
			cond:    func() (bool, error) { time.Sleep(300 * time.Millisecond); return true, nil },
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tk := &Ticker{client: &fakeTickerClient{blocks: map[string]*blockInfo{"best": {Number: 1}}}}
			err := tk.WaitForCondition(tc.timeout, tc.cond)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
