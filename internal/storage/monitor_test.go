package storage

import (
	"errors"
	"math"
	"testing"
)

func TestMonitorRejectsCapacityBelowMinimum(t *testing.T) {
	monitor, err := NewMonitor(t.TempDir(), math.MaxInt64, 20)
	if err != nil {
		t.Fatal(err)
	}
	capacity, err := monitor.Check()
	if err != nil {
		t.Fatal(err)
	}
	if capacity.Available {
		t.Fatalf("capacity unexpectedly available: %#v", capacity)
	}
	if capacity.RequiredFreeBytes != math.MaxInt64 {
		t.Fatalf("required bytes = %d", capacity.RequiredFreeBytes)
	}
	if monitor.Health() == nil {
		t.Fatal("health unexpectedly passed")
	}
}

func TestMonitorUsesPercentageThreshold(t *testing.T) {
	monitor, err := NewMonitor(t.TempDir(), 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	capacity, err := monitor.Check()
	if err != nil {
		t.Fatal(err)
	}
	if capacity.RequiredFreeBytes != capacity.TotalBytes/5 {
		t.Fatalf("capacity = %#v", capacity)
	}
}

func TestMonitorTracksAndReleasesReservations(t *testing.T) {
	monitor, err := NewMonitor(t.TempDir(), 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	capacity, err := monitor.Check()
	if err != nil {
		t.Fatal(err)
	}
	reserved := capacity.FreeBytes - capacity.RequiredFreeBytes
	release, err := monitor.Reserve("job", reserved)
	if err != nil {
		t.Fatal(err)
	}
	capacity, err = monitor.Check()
	if err != nil {
		t.Fatal(err)
	}
	if capacity.ReservedBytes != reserved || !capacity.Available {
		t.Fatalf("capacity = %#v", capacity)
	}
	if _, err := monitor.Reserve("second", 1); !errors.Is(err, ErrInsufficientStorage) {
		t.Fatalf("reserve error = %v", err)
	}
	release()
	capacity, err = monitor.Check()
	if err != nil {
		t.Fatal(err)
	}
	if capacity.ReservedBytes != 0 {
		t.Fatalf("capacity = %#v", capacity)
	}
}
