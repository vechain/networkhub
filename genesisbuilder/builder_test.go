package genesisbuilder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_Delay(t *testing.T) {
	// Case 1: No delay (default)
	b1 := New(1)
	genesis1 := b1.Build()
	now := time.Now().Unix()

	assert.InDelta(t, now, genesis1.LaunchTime, 1, "LaunchTime should be close to now by default")

	// Case 2: With delay
	delay := 10 * time.Second
	b2 := New(1).GenesisTimestampDelay(delay)
	genesis2 := b2.Build()
	expectedTime := time.Now().Add(delay).Unix()

	assert.InDelta(t, expectedTime, genesis2.LaunchTime, 1, "LaunchTime should be delayed by 10 seconds")
}
