package modules

type PodData struct {
	Id string
	// Pod name of the sandbox. Same as the pod name in the Pod ObjectMeta.
	Name string
	// Pod UID of the sandbox. Same as the pod UID in the Pod ObjectMeta.
	Uid string
	// Pod namespace of the sandbox. Same as the pod namespace in the Pod ObjectMeta.
	Namespace string

	ContainerData ContainerData
}

type ContainerData struct {
	// ID of the container.
	Id string
	// Name of the container. Same as the container name in the PodSpec.
	Name string
	// Attempt number of creating the container. Default: 0.
	Attempt uint32

	CreatedAt int64

	StartedAt int64

	FinishedAt int64

	ResourceData ContainerResourceData

	LinuxResourceData LinuxContainerResourceData
}

type ContainerResourceData struct {
	// Cumulative CPU usage (sum across all cores) since object creation.
	CpuUsageCoreNanoSeconds uint64
	// Total CPU usage (sum of all cores) averaged over the sample window.
	// The "core" unit can be interpreted as CPU core-nanoseconds per second.
	CpuUsageNanoCores uint64
	// The amount of working set memory in bytes.
	MemoryWorkingSetBytes uint64
	// Available memory for use. This is defined as the memory limit - workingSetBytes.
	MemoryAvailableBytes uint64
	// Total memory in use. This includes all memory regardless of when it was accessed.
	MemoryUsageBytes uint64
	// The amount of anonymous and swap cache memory (includes transparent hugepages).
	MemoryRssBytes uint64
	// Cumulative number of minor page faults.
	PageFaults uint64
	// Cumulative number of major page faults.
	MajorPageFaults uint64

	//custom resources.
	MemoryUsagePercents float64
}

type LinuxContainerResourceData struct {
	// CPU CFS (Completely Fair Scheduler) period. Default: 0 (not specified).
	CpuPeriod int64
	// CPU CFS (Completely Fair Scheduler) quota. Default: 0 (not specified).
	CpuQuota int64
	// CPU shares (relative weight vs. other containers). Default: 0 (not specified).
	CpuShares int64
	// Memory limit in bytes. Default: 0 (not specified).
	MemoryLimitInBytes int64
	// OOMScoreAdj adjusts the oom-killer score. Default: 0 (not specified).
	OomScoreAdj int64
	// CpusetCpus constrains the allowed set of logical CPUs. Default: "" (not specified).
	CpusetCpus string
	// CpusetMems constrains the allowed set of memory nodes. Default: "" (not specified).
	CpusetMems string
}
