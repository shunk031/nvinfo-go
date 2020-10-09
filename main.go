package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora"
)

// These variables are set in build step
var (
	Version  = "unset"
	Revision = "unset"
)

const (
	memFreeRatio     = 5
	gpuFreeRatio     = 5
	memModerateRatio = 90
	gpuModerateRatio = 75
)

type GpuInfo struct {
	index           int
	gpuUUID         string
	name            string
	memoryUsed      int
	memoryTotal     int
	utilizationGpu  int
	persistanceMode bool
}

type Process struct {
	gpuUUID       string
	pid           int
	usedGpuMemory int
	user          string
	command       string
}

func GetUserFromPid(pid int) string {
	out, err := exec.Command("ps", "ho", "user", strconv.Itoa(pid)).Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSuffix(string(out), "\n")
}

func GetCommandFromPid(pid int) string {
	out, err := exec.Command("ps", "ho", "command", strconv.Itoa(pid)).Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSuffix(string(out), "\n")
}

func NewGpuInfoFromLine(line string) GpuInfo {
	values := strings.Split(line, ", ")

	index, err := strconv.Atoi(values[0])
	if err != nil {
		log.Fatal(err)
	}
	gpuUUID := values[1]
	name := values[2]
	memoryUsed, err := strconv.Atoi(values[3])
	if err != nil {
		log.Fatal(err)
	}
	memoryTotal, err := strconv.Atoi(values[4])
	if err != nil {
		log.Fatal(err)
	}
	utilizationGpu, err := strconv.Atoi(values[5])
	if err != nil {
		log.Fatal(err)
	}
	persistanceMode := values[6]
	return GpuInfo{
		index:           index,
		gpuUUID:         gpuUUID,
		name:            name,
		memoryUsed:      memoryUsed,
		memoryTotal:     memoryTotal,
		utilizationGpu:  utilizationGpu,
		persistanceMode: persistanceMode == "Enabled",
	}
}

func NewProcessFromLine(line string) Process {
	values := strings.Split(line, ", ")

	gpuUUID := values[0]
	pid, err := strconv.Atoi(values[1])
	if err != nil {
		log.Fatal(err)
	}
	user := GetUserFromPid(pid)
	usedGpuMemory, err := strconv.Atoi(values[2])
	if err != nil {
		log.Fatal(err)
	}
	command := GetCommandFromPid(pid)

	return Process{gpuUUID: gpuUUID, pid: pid, usedGpuMemory: usedGpuMemory, user: user, command: command}

}

func RetrieveGpus() map[string]GpuInfo {
	out, err := exec.Command(
		"/usr/bin/env", "nvidia-smi",
		"--format=csv,noheader,nounits",
		"--query-gpu=index,gpu_uuid,name,memory.used,memory.total,utilization.gpu,persistence_mode").Output()

	if err != nil {
		log.Fatal(err)
	}
	outStr := strings.TrimSuffix(string(out), "\n")
	lines := strings.Split(outStr, "\n")

	gpus := make(map[string]GpuInfo, 10)
	for _, line := range lines {
		gpu := NewGpuInfoFromLine(line)
		gpus[gpu.gpuUUID] = gpu
	}
	return gpus
}

func RetrieveProcesses() []Process {
	out, err := exec.Command(
		"/usr/bin/env", "nvidia-smi",
		"--format=csv,noheader,nounits",
		"--query-compute-apps=gpu_uuid,pid,used_memory",
	).Output()
	if err != nil {
		log.Fatal(err)
	}

	outStr := strings.TrimSuffix(string(out), "\n")
	lines := strings.Split(outStr, "\n")
	if lines[0] == "" {
		return []Process{}
	}

	processes := []Process{}
	for _, line := range lines {
		process := NewProcessFromLine(line)
		processes = append(processes, process)
	}

	return processes

}

func gpuProcessExists(gpu GpuInfo, processes []Process) string {
	for _, process := range processes {
		if gpu.gpuUUID == process.gpuUUID {
			return "RUNNING"
		}
	}
	return "-------"
}

func printProcesses(processes []Process, gpus map[string]GpuInfo) string {
	outputs := []string{}
	for _, process := range processes {
		outputs = append(outputs, fmt.Sprintf("| %3d | %10s | %7d | %5d MiB | %22.22s |",
			gpus[process.gpuUUID].index,
			process.user,
			process.pid,
			process.usedGpuMemory,
			process.command))
	}
	return strings.Join(outputs, "\n")
}

func main() {
	gpus := RetrieveGpus()
	processes := RetrieveProcesses()

	fmt.Println("+----------------------------+------+-------------------+---------+")
	fmt.Println("| GPU                        | %GPU | VRAM              | PROCESS |")
	fmt.Println("|----------------------------+------+-------------------+---------|")

	hack := map[int]string{}
	hackkeys := []int{}
	for k, v := range gpus {
		hack[v.index] = k
		hackkeys = append(hackkeys, v.index)
	}
	sort.Ints(hackkeys)
	for _, v := range hackkeys {
		gpuUUID := hack[v]
		gpu := gpus[gpuUUID]

		usedMem := gpu.memoryUsed
		totalMem := gpu.memoryTotal
		gpuUtil := gpu.utilizationGpu
		memUtil := usedMem / totalMem

		isModerate := false
		isHigh := float32(gpuUtil) >= gpuModerateRatio || float32(memUtil) >= memModerateRatio
		if !isHigh {
			isModerate = float32(gpuUtil) >= gpuFreeRatio || float32(memUtil) >= memFreeRatio
		}

		colorFormat := "| %3d %22s | %3d  | %5d / %5d MiB | %s |"
		var auroraFormat aurora.Value
		if isHigh {
			auroraFormat = aurora.Red(colorFormat)
		} else if isModerate {
			auroraFormat = aurora.Yellow(colorFormat)
		} else {
			auroraFormat = aurora.Green(colorFormat)
		}

		output := aurora.Sprintf(
			auroraFormat,
			gpu.index,
			gpu.name,
			gpu.utilizationGpu,
			gpu.memoryUsed,
			gpu.memoryTotal,
			gpuProcessExists(gpu, processes))
		fmt.Println(output)
	}
	fmt.Println("|=================================================================|")

	if len(processes) == 0 {
		fmt.Println("| No running processes found                                      |")
		fmt.Println("+-----------------------------------------------------------------+")
		os.Exit(0)
	}

	fmt.Println("| GPU | USER       | PID     | VRAM      | COMMAND                |")
	fmt.Println("|-----+------------+---------+-----------+------------------------|")
	fmt.Println(printProcesses(processes, gpus))
	fmt.Println("+-----+------------+---------+-----------+------------------------+")

}
