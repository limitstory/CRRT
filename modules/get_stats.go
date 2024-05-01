package modules

import (
	"context"
	"fmt"
	"os"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	internalapi "k8s.io/cri-api/pkg/apis"
	pb "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func PodInfoInit() []PodData {
	var podInfoSet = make([]PodData, 0) //store podInfo set into dynamic array

	return podInfoSet

}

func GetSystemStatsInfo() ([]Percpu, Totalcpu, Memory) {
	var get_per_cpu_set = make([]Percpu, 0) //Dynamic array to store container system metric
	var get_total_cpu Totalcpu
	var get_memory Memory

	per_cpu, err := cpu.Times(true)
	if err != nil {
		panic(err)
	}
	// fmt.Println(per_cpu)

	for i := 0; i < len(per_cpu); i++ {
		var get_per_cpu Percpu

		get_per_cpu.CPU = per_cpu[i].CPU
		get_per_cpu.User = per_cpu[i].User
		get_per_cpu.System = per_cpu[i].System
		get_per_cpu.Idle = per_cpu[i].Idle
		get_per_cpu.Nice = per_cpu[i].Nice
		get_per_cpu.Iowait = per_cpu[i].Iowait
		get_per_cpu.Irq = per_cpu[i].Irq
		get_per_cpu.Softirq = per_cpu[i].Softirq
		get_per_cpu.Steal = per_cpu[i].Steal
		get_per_cpu.Guest = per_cpu[i].Guest
		get_per_cpu.GuestNice = per_cpu[i].GuestNice

		get_per_cpu_set = append(get_per_cpu_set, get_per_cpu) // append to dynamic array
	}

	for i := 0; i < len(get_per_cpu_set); i++ {
		if i == 0 {
			get_total_cpu.CPU = per_cpu[i].CPU
			get_total_cpu.User = per_cpu[i].User
			get_total_cpu.System = per_cpu[i].System
			get_total_cpu.Idle = per_cpu[i].Idle
			get_total_cpu.Nice = per_cpu[i].Nice
			get_total_cpu.Iowait = per_cpu[i].Iowait
			get_total_cpu.Irq = per_cpu[i].Irq
			get_total_cpu.Softirq = per_cpu[i].Softirq
			get_total_cpu.Steal = per_cpu[i].Steal
			get_total_cpu.Guest = per_cpu[i].Guest
			get_total_cpu.GuestNice = per_cpu[i].GuestNice
		} else {
			get_total_cpu.User += per_cpu[i].User
			get_total_cpu.System += per_cpu[i].System
			get_total_cpu.Idle += per_cpu[i].Idle
			get_total_cpu.Nice += per_cpu[i].Nice
			get_total_cpu.Iowait += per_cpu[i].Iowait
			get_total_cpu.Irq += per_cpu[i].Irq
			get_total_cpu.Softirq += per_cpu[i].Softirq
			get_total_cpu.Steal += per_cpu[i].Steal
			get_total_cpu.Guest += per_cpu[i].Guest
			get_total_cpu.GuestNice += per_cpu[i].GuestNice
		}
	}
	get_total_cpu.CPU = "cpu-total"

	// fmt.Println(get_total_cpu)

	memory, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}
	// fmt.Println(memory)

	get_memory.Total = memory.Total
	get_memory.Available = memory.Available
	get_memory.Used = memory.Total - memory.Available
	get_memory.UsedPercent = float64(get_memory.Used) / float64(memory.Total) * 100

	return get_per_cpu_set, get_total_cpu, get_memory
}

func GetListPodStatsInfo(client internalapi.RuntimeService) []*pb.PodSandboxStats {
	for {
		filter := &pb.PodSandboxStatsFilter{}

		stats, err := client.ListPodSandboxStats(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		} else {
			return stats
		}
	}
}

func GetPodStatsInfo(client internalapi.RuntimeService, podInfoSet []PodData) []PodData {

	stats := GetListPodStatsInfo(client)

	for i := 0; i < len(stats); i++ {

		// Do not store namespaces other than default namespaces
		if stats[i].Attributes.Metadata.Namespace != "default" {
			continue
		}

		// Do not store info of notworking pods
		status, err := client.PodSandboxStatus(context.TODO(), stats[i].Attributes.Id, false)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if status.Status.State == 1 { // exception handling: SANDBOX_NOTREADY
			continue
		}

		var podInfo PodData

		podInfo.Id = stats[i].Attributes.Id
		podInfo.Name = stats[i].Attributes.Metadata.Name
		podInfo.Uid = stats[i].Attributes.Metadata.Uid
		podInfo.Namespace = stats[i].Attributes.Metadata.Namespace

		// 현재 파드당 싱글 컨테이너만 있다고 가정하고 코드 작성
		// 파드에 여러개 컨테이너가 들어있을 수 있어 구조체에 동적 배열 도입 및 로직 처리 필요
		if len(stats[i].Linux.Containers) == 0 { // exception handling: pod is created, but it not have containers
			continue
		}
		podInfo.ContainerData.Id = stats[i].Linux.Containers[0].Attributes.Id
		podInfo.ContainerData.Name = stats[i].Linux.Containers[0].Attributes.Metadata.Name
		podInfo.ContainerData.Attempt = stats[i].Linux.Containers[0].Attributes.Metadata.Attempt

		if stats[i].Linux.Containers[0].Cpu == nil { // exception handling: nil pointer
			continue
		}
		if stats[i].Linux.Containers[0].Memory == nil { // exception handling: nil pointer
			continue
		}
		podInfo.ContainerData.ResourceData.CpuUsageCoreNanoSeconds = stats[i].Linux.Containers[0].Cpu.UsageCoreNanoSeconds.Value
		podInfo.ContainerData.ResourceData.CpuUsageNanoCores = stats[i].Linux.Containers[0].Cpu.UsageNanoCores.Value

		podInfo.ContainerData.ResourceData.MemoryAvailableBytes = stats[i].Linux.Containers[0].Memory.AvailableBytes.Value
		podInfo.ContainerData.ResourceData.MemoryWorkingSetBytes = stats[i].Linux.Containers[0].Memory.WorkingSetBytes.Value
		podInfo.ContainerData.ResourceData.MemoryUsageBytes = stats[i].Linux.Containers[0].Memory.UsageBytes.Value

		podInfoSet = append(podInfoSet, podInfo) // append to dynamic array
	}

	if len(podInfoSet) == 0 {
		fmt.Println("There is no pod running.")
		os.Exit(0)
	}

	return podInfoSet
}

func GetContainerStatsInfo(client internalapi.RuntimeService, podInfoSet []PodData, resource []*pb.ContainerResources) ([]PodData, []*pb.ContainerResources) {
	for i := 0; i < len(podInfoSet); i++ {
		containerStats, err := client.ContainerStatus(context.TODO(), podInfoSet[i].ContainerData.Id, false)
		if err != nil { // exception handling
			fmt.Println(err)
			fmt.Println("Remove Pod Set")
			podInfoSet = RemovePodofPodInfoSet(podInfoSet, i)
			i--
		}

		podInfoSet[i].ContainerData.LinuxResourceData.CpuPeriod = containerStats.Status.Resources.Linux.CpuPeriod
		podInfoSet[i].ContainerData.LinuxResourceData.CpuQuota = containerStats.Status.Resources.Linux.CpuQuota
		podInfoSet[i].ContainerData.LinuxResourceData.CpuShares = containerStats.Status.Resources.Linux.CpuShares
		podInfoSet[i].ContainerData.LinuxResourceData.MemoryLimitInBytes = containerStats.Status.Resources.Linux.MemoryLimitInBytes
		podInfoSet[i].ContainerData.LinuxResourceData.OomScoreAdj = containerStats.Status.Resources.Linux.OomScoreAdj
		podInfoSet[i].ContainerData.LinuxResourceData.CpusetCpus = containerStats.Status.Resources.Linux.CpusetCpus
		podInfoSet[i].ContainerData.LinuxResourceData.CpusetMems = containerStats.Status.Resources.Linux.CpusetMems

		podInfoSet[i].ContainerData.CreatedAt = containerStats.Status.CreatedAt
		podInfoSet[i].ContainerData.StartedAt = containerStats.Status.StartedAt
		podInfoSet[i].ContainerData.FinishedAt = containerStats.Status.FinishedAt

		resource = append(resource, containerStats.Status.Resources) // append to dynamic array
	}
	return podInfoSet, resource
}
