package identity

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemInfo struct {
	Hostname    string
	OS          string
	Platform    string
	Arch        string
	CpuCores    int32
	TotalMemory int64
	Version     string
	DiskSize    int64
}

func Collect(version string) (*SystemInfo, error) {
	info, err := host.Info()
	if err != nil {
		return nil, err
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	diskUsage, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return &SystemInfo{
		Hostname:    info.Hostname,
		OS:          info.OS,
		Platform:    info.Platform,
		Arch:        info.KernelArch,
		CpuCores:    int32(runtime.NumCPU()),
		TotalMemory: int64(vm.Total),
		Version:     version,
		DiskSize:    int64(diskUsage.Total),
	}, nil
}

func LoadOrCreateAgentID() (string, error) {
	path := filepath.Join(".", ".kanshi-id")

	// If exists, read it
	if data, err := os.ReadFile(path); err == nil {
		return string(data), nil
	}

	// Otherwise generate new one
	id := uuid.NewString()

	if err := os.WriteFile(path, []byte(id), 0644); err != nil {
		return "", err
	}

	return id, nil
}
