package storage

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

type Capacity struct {
	TotalBytes        uint64 `json:"totalBytes"`
	FreeBytes         uint64 `json:"freeBytes"`
	RequiredFreeBytes uint64 `json:"requiredFreeBytes"`
	Available         bool   `json:"available"`
}

type Monitor struct {
	dataDir        string
	minFreeBytes   uint64
	minFreePercent uint64
}

func NewMonitor(dataDir string, minFreeBytes int64, minFreePercent int) (*Monitor, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	return &Monitor{
		dataDir:        dataDir,
		minFreeBytes:   uint64(max(minFreeBytes, 0)),
		minFreePercent: uint64(min(max(minFreePercent, 0), 100)),
	}, nil
}

func (m *Monitor) Name() string { return "disk" }

func (m *Monitor) Check() (Capacity, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(m.dataDir, &stat); err != nil {
		return Capacity{}, err
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	required := max(m.minFreeBytes, total*m.minFreePercent/100)
	return Capacity{
		TotalBytes: total, FreeBytes: free, RequiredFreeBytes: required, Available: free >= required,
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
