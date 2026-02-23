package identity

import (
	"os"
	"runtime"

	"github.com/shirou/gopsutil/v4/mem"
)

type SystemInfo struct {
	Hostname    string
	OS          string
	Arch        string
	CpuCores    int32
	TotalMemory int64
	Version     string
}

func Collect(version string) (*SystemInfo, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return &SystemInfo{
		Hostname:    host,
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		CpuCores:    int32(runtime.NumCPU()),
		TotalMemory: int64(vm.Total),
		Version:     version,
	}, nil
}
