package main

import (
	"time"

	pb "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/kubernetes/pkg/kubelet/cri/remote"

	mod "memory/2gb/modules"
)

func main() {
	const ENDPOINT string = "unix:///var/run/containerd/containerd.sock"
	const MEMORY_LIMIT_THRESHOLD float64 = 0.2
	const DEFAULT_CPU_QUOTA int64 = 200000

	/*
		// kubernetes api 클라이언트 생성하는 모듈
		clientset := mod.InitClient()
		if clientset != nil {
			fmt.Println("123")
		}*/

	//get new internal client service
	client, err := remote.NewRemoteRuntimeService(ENDPOINT, time.Second*2, nil)
	if err != nil {
		panic(err)
	}
	// remote.NewRemoteImageService("unix:///var/run/containerd/containerd.sock", time.Second*2, nil)

	// execute monitoring & resource management logic
	for {
		// definition of data structure to store
		var selectContainerId = make([]string, 0)
		var selectContainerResource = make([]*pb.ContainerResources, 0)

		// get system metrics & memory usage exceeds to threshold value
		timeout := mod.MonitoringSystemResources(false)
		if timeout != 0 {
			panic("err")
		}

		selectContainerId, selectContainerResource = mod.LimitContainerResources(client, selectContainerId, selectContainerResource)

		// After limiting CPU usage, watch the trend of memory usage.
		mod.ControlRecursiveContainerResources(client, selectContainerId, selectContainerResource)

		// 단순히 퍼센트만 가지고 따지면 문제가 발생할 수 있음.
		// ex) 1G에 100m를 사용하고 있는 것보다 전체 10G에 1.2G 사용하고 있는 것을 제한하는 것이 더 효과적일 수 있음.
		// 이를 수학적으로 모델링할 필요성이 있음.
	}
}
