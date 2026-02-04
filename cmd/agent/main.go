package main

import (
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
)

func main() {
	percent, _ := cpu.Percent(0, false)
	fmt.Println(int(percent[0]))
}
