package modules

import (
	"context"
	"fmt"
	"time"

	internalapi "k8s.io/cri-api/pkg/apis"
	pb "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func UpdateContainerResources(client internalapi.RuntimeService, id string, resource *pb.ContainerResources) {

	err := client.UpdateContainerResources(context.TODO(), id, resource)
	if err != nil {
		fmt.Println(err)
	}
}

func MonitoringSystemResources(reverse bool) int32 {
	var round int32 = 0

	for {
		// get system metrics
		per_cpu, total_cpu, memory := GetSystemStatsInfo()
		if per_cpu == nil {
			fmt.Println(per_cpu)
			fmt.Println(total_cpu)
		}

		if reverse == false {
			// if memory usage exceeds to threshold value
			if memory.UsedPercent > MEMORY_THRESHOLD {
				return 0
			}
			fmt.Printf("Total: %d, Available:%d, Used:%d, UsedPercent:%f%% \n", memory.Total, memory.Available, memory.Used, memory.UsedPercent)
			time.Sleep(1 * time.Second)
		} else {
			// if memory usage not exceeds to threshold value
			if memory.UsedPercent < MEMORY_LIMIT_THRESHOLD {
				return 0
			}
			if round >= TIMEOUT_INTERVAL {
				return 1
			}
			fmt.Printf("Total: %d, Available:%d, Used:%d, UsedPercent:%f%% \n", memory.Total, memory.Available, memory.Used, memory.UsedPercent)
			time.Sleep(1 * time.Second)
			round++
		}
	}
}

func MonitoringPodResources(client internalapi.RuntimeService) ([]PodData, []*pb.ContainerResources) {
	var containerResourceSet = make([]*pb.ContainerResources, 0) //Dynamic array to store container system metric

	// 우선 call by reference 방식이 아닌 call by value 방식으로 구현 및 작동 확인함. 추후 공부 후 call by reference 방식으로 변경 필요
	podInfoSet := PodInfoInit()
	// get pod stats
	podInfoSet = GetPodStatsInfo(client, podInfoSet)

	// get container stats
	podInfoSet, containerResourceSet = GetContainerStatsInfo(client, podInfoSet, containerResourceSet)

	// get memory usage percents each containers
	podInfoSet = GetmemoryUsagePercents(podInfoSet)

	return podInfoSet, containerResourceSet
}

func RemoveContainer(client internalapi.RuntimeService, selectContainerId []string, selectContainerResource []*pb.ContainerResources) ([]string, []*pb.ContainerResources) {
	err := client.RemoveContainer(context.TODO(), selectContainerId[len(selectContainerId)-1])
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println()
	fmt.Println("Remove Container Id:", selectContainerId[len(selectContainerId)-1])

	return selectContainerId[:len(selectContainerId)-1], selectContainerResource[:len(selectContainerResource)-1]
}

func LimitContainerResources(client internalapi.RuntimeService, selectContainerId []string, selectContainerResource []*pb.ContainerResources) ([]string, []*pb.ContainerResources) {

	var selectMemoryUsagePercents float64
	var indexOfSelectContainers int32

	// Monitoring Pod Resources
	podInfoSet, containerResourceSet := MonitoringPodResources(client)

	// select restrict containers
	selectMemoryUsagePercents, indexOfSelectContainers, selectContainerId = SelectRestrictContainers(podInfoSet, selectContainerId)

	// kill the last restricted container because all containers are restrict
	// update selectContainerId and selectContainerResource
	if selectMemoryUsagePercents == 100.00 {
		RemoveContainer(client, selectContainerId, selectContainerResource)

		return selectContainerId, selectContainerResource
	}

	// append select containers
	selectContainerId = append(selectContainerId, podInfoSet[indexOfSelectContainers].ContainerData.Id)
	selectContainerResource = append(selectContainerResource, containerResourceSet[indexOfSelectContainers])

	fmt.Printf("\nRestrict ContainerId:%s, UsedPercent:%f%%\n", podInfoSet[indexOfSelectContainers].ContainerData.Id, selectMemoryUsagePercents*100)

	// limit CPU usage for containers with the low memory usage percents
	selectContainerResource[len(selectContainerResource)-1].Linux.CpuQuota = LIMIT_CPU_QUOTA // limit cpu usage to 10m
	UpdateContainerResources(client, selectContainerId[len(selectContainerId)-1], selectContainerResource[len(selectContainerResource)-1])

	// select restrict containers
	selectMemoryUsagePercents, indexOfSelectContainers, selectContainerId = SelectRestrictContainers(podInfoSet, selectContainerId)

	// kill the last restricted container because all containers are restrict
	// update selectContainerId and selectContainerResource
	if selectMemoryUsagePercents == 100.00 {
		RemoveContainer(client, selectContainerId, selectContainerResource)

		return selectContainerId, selectContainerResource
	}

	// append select containers
	selectContainerId = append(selectContainerId, podInfoSet[indexOfSelectContainers].ContainerData.Id)
	selectContainerResource = append(selectContainerResource, containerResourceSet[indexOfSelectContainers])

	fmt.Printf("\nRestrict ContainerId:%s, UsedPercent:%f%%\n", podInfoSet[indexOfSelectContainers].ContainerData.Id, selectMemoryUsagePercents*100)

	// limit CPU usage for containers with the low memory usage percents
	selectContainerResource[len(selectContainerResource)-1].Linux.CpuQuota = LIMIT_CPU_QUOTA // limit cpu usage to 10m
	UpdateContainerResources(client, selectContainerId[len(selectContainerId)-1], selectContainerResource[len(selectContainerResource)-1])

	return selectContainerId, selectContainerResource
}

func ControlRecursiveContainerResources(client internalapi.RuntimeService, selectContainerId []string, selectContainerResource []*pb.ContainerResources) {
	// get system metrics & memory usage exceeds to threshold value
	timeout := MonitoringSystemResources(true)

	if timeout == 1 { // when timeout occurs, additional containers are restricted and monitor system resource again
		selectContainerId, selectContainerResource = LimitContainerResources(client, selectContainerId, selectContainerResource)
		// recursive하게 호출
		ControlRecursiveContainerResources(client, selectContainerId, selectContainerResource)
	} else { // revert CPU usage of all containers if memory usage is low
		for i := 0; i < len(selectContainerId); i++ {
			selectContainerResource[i].Linux.CpuQuota = DEFAULT_CPU_QUOTA
			UpdateContainerResources(client, selectContainerId[i], selectContainerResource[i])
			fmt.Printf("\nRelease ContainerId:%s\n", selectContainerId[i])
		}
	}
}
