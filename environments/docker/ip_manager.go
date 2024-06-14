package docker

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type IpManager struct {
	baseIP      string
	currentIP   int
	assignedIps map[string]string
}

func NewIPManagerRandom() *IpManager {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	// Generate random octets for a Class C subnet
	a := rand.Intn(128)
	b := rand.Intn(256)
	c := rand.Intn(256)
	return NewIPManager(fmt.Sprintf("%d.%d.%d.0", a, b, c))
}

func NewIPManager(subnet string) *IpManager {
	// Start from 2 because .2 is the first available IP
	return &IpManager{
		baseIP:      subnet[:len(subnet)-1], // Remove the trailing zero
		currentIP:   2,
		assignedIps: map[string]string{},
	}
}

func (im *IpManager) NextIP(nodeID string) (string, error) {
	if im.currentIP > 253 {
		return "", errors.New("no more available IP addresses")
	}
	ipAddr := fmt.Sprintf("%s%d", im.baseIP, im.currentIP)
	im.currentIP++
	im.assignedIps[nodeID] = ipAddr
	return ipAddr, nil
}

func (im *IpManager) Subnet() string {
	return fmt.Sprintf("%s0/24", im.baseIP)
}

func (im *IpManager) GetNodeIP(nodeID string) string {
	return im.assignedIps[nodeID]
}
