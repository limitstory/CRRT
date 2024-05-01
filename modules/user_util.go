package modules

func GetmemoryUsagePercents(podInfoSet []PodData) []PodData {
	// get current container memory usage and limit value
	for i := 0; i < len(podInfoSet); i++ {
		containerMemoryUsages := podInfoSet[i].ContainerData.ResourceData.MemoryUsageBytes
		// if limit is not set, it will appear as 0; if set, it will output normally.
		containerMemoryLimits := podInfoSet[i].ContainerData.LinuxResourceData.MemoryLimitInBytes

		// exception handling
		// container without limit set, not burstable container
		if containerMemoryLimits == 0 {
			podInfoSet[i].ContainerData.ResourceData.MemoryUsagePercents = 0
		}
		podInfoSet[i].ContainerData.ResourceData.MemoryUsagePercents = float64(containerMemoryUsages) / float64(containerMemoryLimits)
	}

	return podInfoSet
}

func SelectRestrictContainers(podInfoSet []PodData, selectContainerId []string) (float64, int32, []string) {

	selectMemoryUsagePercents := 100.00
	indexOfSelectContainers := 0

	for i := 0; i < len(podInfoSet); i++ {
		// container without limit set, not burstable container (exception handling)
		// not lowest memory usage percents
		if podInfoSet[i].ContainerData.ResourceData.MemoryUsagePercents == 0 ||
			selectMemoryUsagePercents < podInfoSet[i].ContainerData.ResourceData.MemoryUsagePercents {
			continue
		}

		// verify if already restricted the container resources
		isOverlap := false
		for j := 0; j < len(selectContainerId); j++ {
			if podInfoSet[i].ContainerData.Id == selectContainerId[j] {
				isOverlap = true
				break
			}
		}
		if isOverlap == true {
			continue
		}

		// choose lowest memory usage percents
		selectMemoryUsagePercents = podInfoSet[i].ContainerData.ResourceData.MemoryUsagePercents
		indexOfSelectContainers = i
	}

	return selectMemoryUsagePercents, int32(indexOfSelectContainers), selectContainerId
}

func RemovePodofPodInfoSet(podInfoSet []PodData, i int) []PodData {
	podInfoSet[i] = podInfoSet[len(podInfoSet)-1]
	return podInfoSet[:len(podInfoSet)-1]
}
