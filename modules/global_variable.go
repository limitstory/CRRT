package modules

const ENDPOINT string = "unix:///var/run/containerd/containerd.sock"
const MEMORY_THRESHOLD float64 = 94.0
const MEMORY_LIMIT_THRESHOLD float64 = 90.0
const DEFAULT_CPU_QUOTA int64 = 200000
const LIMIT_CPU_QUOTA int64 = 2000

const TIMEOUT_INTERVAL int32 = 3
