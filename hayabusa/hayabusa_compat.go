package hayabusa

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/thorclient"
)

// Method selectors for staker contract methods
const (
	FirstActiveSelector = "0xd719835c" // firstActive()
)

// Staker provides compatibility interface for staker contract interactions
type Staker struct {
	client     *thorclient.Client
	stakerAddr thor.Address
}

// NewStaker creates a new staker instance
func NewStaker(client *thorclient.Client) *Staker {
	return &Staker{
		client:     client,
		stakerAddr: thor.BytesToAddress([]byte("Staker")),
	}
}

// FirstActive calls the firstActive method on the staker contract
func (s *Staker) FirstActive() (thor.Address, error) {
	payload := fmt.Sprintf(`{"clauses":[{"to":"%s","value":"0x0","data":"%s"}]}`,
		s.stakerAddr.String(), FirstActiveSelector)

	result, statusCode, err := s.client.RawHTTPClient().RawHTTPPost("/accounts/*", []byte(payload))
	if err != nil {
		return thor.Address{}, fmt.Errorf("failed to call firstActive: %w", err)
	}

	if statusCode != 200 {
		return thor.Address{}, fmt.Errorf("firstActive call failed with status %d", statusCode)
	}

	// Parse the response
	var response []struct {
		Data     string `json:"data"`
		Reverted bool   `json:"reverted"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return thor.Address{}, fmt.Errorf("failed to parse firstActive response: %w", err)
	}

	if len(response) == 0 || response[0].Reverted {
		return thor.Address{}, fmt.Errorf("firstActive call reverted")
	}

	// Decode ABI-encoded address
	data := response[0].Data
	if len(data) < 2 || data[:2] != "0x" {
		return thor.Address{}, fmt.Errorf("invalid response data format: %s", data)
	}

	dataBytes, err := hex.DecodeString(data[2:])
	if err != nil {
		return thor.Address{}, fmt.Errorf("failed to decode response data: %w", err)
	}

	if len(dataBytes) < 32 {
		return thor.Address{}, fmt.Errorf("unexpected return length: %d", len(dataBytes))
	}

	// Last 20 bytes of the 32-byte word are the address
	var validatorAddr thor.Address
	copy(validatorAddr[:], dataBytes[12:32])

	return validatorAddr, nil
}
