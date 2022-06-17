package main

import (
	"flag"
	"fmt"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"syscall"
)

var (
	namespace = "default"
)

func main() {
	outsideCluster := flag.Bool("outside-cluster", false, "set to true when run out of cluster. (default: false)")
	flag.Parse()

	clientset := GetClientSet(*outsideCluster)

	controller := NewServiceController(clientset)
	stop := make(chan struct{})
	defer close(stop)
	err := controller.Run(stop)
	if err != nil {
		klog.Fatal(err)
	}

	fmt.Println("Waiting for Kill Signal...")
	var stopChan = make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-stopChan
}
