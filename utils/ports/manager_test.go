package ports

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	m := &Manager{ports: make(map[int]bool), portsByID: make(map[string][]int)}
	// reset global default to avoid cross-test contamination
	defaultManager = m
	return m
}

func TestManager_AllocateAndRelease(t *testing.T) {
	m := newTestManager(t)

	p1, err := m.Allocate("net-1")
	require.NoError(t, err, "allocate failed")
	require.NotZero(t, p1, "expected non-zero port")

	p2, err := m.Allocate("net-1")
	require.NoError(t, err, "allocate failed")
	require.NotEqual(t, p1, p2, "expected unique ports")

	require.NoError(t, m.ReleaseAll("net-1"), "release failed")

	// Verify we can re-allocate after release (ports are freed)
	_, err = m.Allocate("net-1")
	require.NoError(t, err, "re-allocate after release failed")
}

func TestManager_Allocate_ErrOnEmptyID(t *testing.T) {
	m := newTestManager(t)
	_, err := m.Allocate("")
	require.Error(t, err, "expected error for empty network id")
}

func TestManager_ReleaseAll_ErrOnEmptyID(t *testing.T) {
	m := newTestManager(t)
	err := m.ReleaseAll("")
	require.Error(t, err, "expected error for empty network id")
}

func TestManager_AllocateUniqueAcrossIDs(t *testing.T) {
	m := newTestManager(t)

	const n = 20
	seen := make(map[int]bool)

	for range n {
		p, err := m.Allocate("A")
		require.NoError(t, err, "alloc A failed")
		require.False(t, seen[p], "duplicate port across ids not expected yet")
		seen[p] = true
	}
	for range n {
		p, err := m.Allocate("B")
		require.NoError(t, err, "alloc B failed")
		require.False(t, seen[p], "duplicate port across different ids")
		seen[p] = true
	}
}

func TestManager_ConcurrentAllocationsUnique(t *testing.T) {
	m := newTestManager(t)

	const workers = 64
	var wg sync.WaitGroup
	wg.Add(workers)

	mu := sync.Mutex{}
	seen := make(map[int]bool)

	for range workers {
		go func() {
			defer wg.Done()
			p, err := m.Allocate("C")
			if err != nil {
				t.Errorf("alloc failed: %v", err)
				return
			}
			mu.Lock()
			if seen[p] {
				t.Errorf("duplicate port detected: %d", p)
			}
			seen[p] = true
			mu.Unlock()
		}()
	}
	wg.Wait()

	require.Equal(t, workers, len(seen), "expected %d unique ports, got %d", workers, len(seen))
}

func TestManager_ReleaseAll_AllowsBulkRealloc(t *testing.T) {
	m := newTestManager(t)

	const n = 40
	for range n {
		_, err := m.Allocate("D")
		require.NoError(t, err, "alloc failed")
	}
	require.NoError(t, m.ReleaseAll("D"), "release failed")
	// Should be able to re-allocate n ports again without collisions
	seen := make(map[int]bool)
	for range n {
		p, err := m.Allocate("D")
		require.NoError(t, err, "re-alloc failed")
		require.False(t, seen[p], "duplicate port upon re-alloc")
		seen[p] = true
	}
}
