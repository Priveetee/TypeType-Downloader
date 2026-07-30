package storage

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

var ErrInsufficientStorage = errors.New("insufficient storage")

type Capacity struct {
	TotalBytes        uint64 `json:"totalBytes"`
	FreeBytes         uint64 `json:"freeBytes"`
	RequiredFreeBytes uint64 `json:"requiredFreeBytes"`
	ReservedBytes     uint64 `json:"reservedBytes"`
	Available         bool   `json:"available"`
}

type Monitor struct {
	mu             sync.Mutex
	dataDir        string
	minFreeBytes   uint64
	minFreePercent uint64
	reservations   map[string]uint64
}

func NewMonitor(dataDir string, minFreeBytes int64, minFreePercent int) (*Monitor, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	return &Monitor{
		dataDir:        dataDir,
		minFreeBytes:   uint64(max(minFreeBytes, 0)),
		minFreePercent: uint64(min(max(minFreePercent, 0), 100)),
		reservations:   make(map[string]uint64),
	}, nil
}

func (m *Monitor) Name() string { return "disk" }

func (m *Monitor) Check() (Capacity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.check()
}

func (m *Monitor) Reserve(id string, bytes uint64) (func(), error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	capacity, err := m.check()
	if err != nil {
		return nil, err
	}
	previous := m.reservations[id]
	reserved := capacity.ReservedBytes - previous
	required := saturatedAdd(capacity.RequiredFreeBytes-previous, bytes)
	if capacity.FreeBytes < required {
		return nil, fmt.Errorf(
			"%w: need %d bytes, have %d bytes with %d bytes reserved",
			ErrInsufficientStorage,
			required,
			capacity.FreeBytes,
			reserved,
		)
	}
	m.reservations[id] = bytes
	return func() {
		m.mu.Lock()
		delete(m.reservations, id)
		m.mu.Unlock()
	}, nil
}

func (m *Monitor) check() (Capacity, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(m.dataDir, &stat); err != nil {
		return Capacity{}, err
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	reserved := uint64(0)
	for _, bytes := range m.reservations {
		reserved = saturatedAdd(reserved, bytes)
	}
	minimum := max(m.minFreeBytes, total*m.minFreePercent/100)
	required := saturatedAdd(minimum, reserved)
	return Capacity{
		TotalBytes: total, FreeBytes: free, RequiredFreeBytes: required,
		ReservedBytes: reserved, Available: free >= required,
	}, nil
}

func (m *Monitor) Health() error {
	capacity, err := m.Check()
	if err != nil {
		return err
	}
	if capacity.Available {
		return nil
	}
	return fmt.Errorf("free bytes %d below required %d", capacity.FreeBytes, capacity.RequiredFreeBytes)
}

func saturatedAdd(left uint64, right uint64) uint64 {
	if ^uint64(0)-left < right {
		return ^uint64(0)
	}
	return left + right
}
