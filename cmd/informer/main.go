package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	namespace     = "default"
	clientset     *kubernetes.Clientset
	newdeployment *appv1.Deployment
)

func main() {
	outsideCluster := flag.Bool("outside-cluster", false, "set to true when run out of cluster. (default: false)")
	flag.Parse()

	if *outsideCluster {
		// creates the out-cluster config
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		config, err := clientcmd.BuildConfigFromFlags("", path.Join(home, ".kube/config"))
		if err != nil {
			panic(err.Error())
		}
		// creates the clientset
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
	} else {
		// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		// creates the clientset
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
	}
	controller := NewDeploymentController(clientset)

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
func int32Ptr(i int32) *int32 { return &i }

func getService(client kubernetes.Interface, namespace, name string) (*corev1.Service, error) {
	cm, err := client.
		CoreV1().
		Services(namespace).
		Get(
			context.Background(),
			name,
			metav1.GetOptions{},
		)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func deleteService(client kubernetes.Interface, namespace, name string) error {
	err := client.
		CoreV1().
		Services(namespace).
		Delete(
			context.Background(),
			name,
			metav1.DeleteOptions{},
		)
	return err
}

var portnum int32 = 80

func createService(client kubernetes.Interface, namespace, name string, newdeployment *appv1.Deployment) (*corev1.Service, error) {
	cm := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sm-service",
			Labels: map[string]string{
				"ntcu-k8s": "hw3",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: newdeployment.Spec.Selector.MatchLabels,
			Type:     corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.IntOrString{IntVal: portnum},
					NodePort:   30100,
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	cm.Namespace = namespace
	cm.Name = name
	cm, err := client.
		CoreV1().
		Services(namespace).
		Create(
			context.Background(),
			cm,
			metav1.CreateOptions{},
		)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
