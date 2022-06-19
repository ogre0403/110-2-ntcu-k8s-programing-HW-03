package main

import (
	"context"
	"fmt"
	"time"

	"os"
	"path"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/informers"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type DeploymentController struct {
	informerFactory informers.SharedInformerFactory
	informer        appsinformer.DeploymentInformer
	clientSet       *kubernetes.Clientset
	svc              *corev1.Service
	dep              *appv1.Deployment
}

var clientSet *kubernetes.Clientset


func (c *DeploymentController) Run(stopCh chan struct{}) error {
	// Starts all the shared informers that have been created by the factory so
	// far.
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.informer.Informer().HasSynced) {
		return fmt.Errorf("Failed to sync")
	}
	return nil
}

func (c *DeploymentController) onAdd(obj interface{}) {
	job := obj.(*appv1.Deployment)
	if !(job.GetLabels()["ntcu-k8s"] == "hw3") {
		return
	}
	c.svc = createService(c.clientSet,job)
	fmt.Printf("Informer event: Job ADDED %s/%s\n", job.GetNamespace(), job.GetName())

}

func (c *DeploymentController) onUpdate(old, new interface{}) {
	job := old.(*appv1.Deployment)
	fmt.Printf("Informer event: Job UPDATED %s/%s\n", job.GetNamespace(), job.GetName())

}

func (c *DeploymentController) onDelete(obj interface{}) {
	job := obj.(*appv1.Deployment)
	fmt.Printf("Informer event: Job DELETED %s/%s\n", job.GetNamespace(), job.GetName())

	if err := deleteService(c.clientSet,c.svc); err == nil {
		fmt.Printf("----Delete service when Job DELETE Event %s/%s\n", c.svc.GetNamespace(), c.svc.GetName())
	}

}


func NDeploymentController(client *kubernetes.Clientset) *DeploymentController {
	factory := informers.NewSharedInformerFactoryWithOptions(client, 5*time.Second, informers.WithNamespace(namespace))
	informer := factory.Apps().V1().Deployments()

	c := &DeploymentController{
		informerFactory: factory,
		informer:        informer,
		clientSet:       client,
	}
	informer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.onAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.onUpdate,
			// Called on resource deletion.
			DeleteFunc: c.onDelete,
		},
	)
	return c
}

func int32Ptr(i int32) *int32 { return &i }


func deleteService(client kubernetes.Interface,svc *corev1.Service) error{
	err := client.
		CoreV1().
		Services(svc.GetNamespace()).
		Delete(
			context.Background(),
			svc.GetName(),
			metav1.DeleteOptions{},
		)
	if err != nil {
		panic(err.Error())
	}

	return err
}

var portnum int32 = 80

func createService(client kubernetes.Interface, dep *appv1.Deployment) *corev1.Service {
	sm := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "service",
			Labels: map[string]string{
				"ntcu-k8s": "hw3",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: dep.Spec.Selector.MatchLabels,
			Type: corev1.ServiceTypeNodePort,
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
	sm.Namespace = namespace
	sm, err := client.
		CoreV1().
		Services(namespace).
		Create(
			context.Background(),
			sm,
			metav1.CreateOptions{},
		)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Created Deplyment %s/%s\n", sm.GetNamespace(), sm.GetName())
	return sm
}



func GetClientSet(isOut bool) *kubernetes.Clientset {
	var clientset *kubernetes.Clientset
	if isOut {
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

	return clientset
}
