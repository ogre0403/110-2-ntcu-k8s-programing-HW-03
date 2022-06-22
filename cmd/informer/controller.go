package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/informers"
	batchinformer "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"
	"context"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ServiceController struct {
	informerFactory informers.SharedInformerFactory
	informer        batchinformer.DeploymentInformer
	clientSet       *kubernetes.Clientset
	svc 			*corev1.Service
	dpm				*appv1.Deployment
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *ServiceController) Run(stopCh chan struct{}) error {
	// Starts all the shared informers that have been created by the factory so
	// far.
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.informer.Informer().HasSynced) {
		return fmt.Errorf("Failed to sync")
	}
	return nil
}

func (c *ServiceController) onAdd(obj interface{}) {
	dmc := obj.(*appv1.Deployment)
	if !(dmc.GetLabels()["ntcu-k8s"] == "hw3") {
		return
	}
	fmt.Printf("Informer event: Deployment ADDED %s/%s\n", dmc.GetNamespace(), dmc.GetName())
}

func (c *ServiceController) onUpdate(old, new interface{}) {
	dmc := old.(*appv1.Deployment)
	if !(dmc.GetLabels()["ntcu-k8s"] == "hw3") {
		return
	}
	_, err := GetService(c.clientSet, namespace, "ntcu-nginx")
	if err != nil && errors.IsNotFound(err) {
		dm, _ := CreateService(c.clientSet, namespace, "ntcu-nginx")
		fmt.Printf("----Create Service when Deployment UPDATED Event %s/%s\n", dm.GetNamespace(), dm.GetName())
	}
	fmt.Printf("Informer event: Deployment UPDATED %s/%s\n", dmc.GetNamespace(), dmc.GetName())
}

func (c *ServiceController) onDelete(obj interface{}) {
	dmc := obj.(*appv1.Deployment)
	if !(dmc.GetLabels()["ntcu-k8s"] == "hw3") {
		return
	}
	fmt.Printf("Informer event: Deployment DELETED %s/%s\n", dmc.GetNamespace(), dmc.GetName())

	if err := DeleteService(c.clientSet, namespace, "ntcu-nginx"); err == nil {
		fmt.Printf("----Delete Service when Deployment DELETE Event %s/%s\n", namespace, "ntcu-nginx")
	}
}

// NewServiceController creates a ServiceController
func NewServiceController(client *kubernetes.Clientset) *ServiceController {
	/*factory := informers.NewSharedInformerFactoryWithOptions(client, 5*time.Second, informers.WithNamespace(namespace), informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.LabelSelector = "ntcu-k8s=hw3"
	}))*/
	factory := informers.NewSharedInformerFactoryWithOptions(client, 5*time.Second, informers.WithNamespace(namespace))
	informer := factory.Apps().V1().Deployments()

	c := &ServiceController{
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

func GetService(client kubernetes.Interface, namespace, name string) (*corev1.Service, error) {
	sm, err := client.
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
	return sm, nil
}

func CreateService(client kubernetes.Interface, namespace, name string) (*corev1.Service, error) {
	sm := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ntcu-nginx",
			Labels: map[string]string{
				"ntcu-k8s": "hw3",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"ntcu-k8s": "hw3",
			},
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.IntOrString{IntVal: portnum},
					NodePort:   30080,
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	sm.Namespace = namespace
	sm.Name = name

	sm, err := client.
		CoreV1().
		Services(namespace).
		Create(
			context.Background(),
			sm,
			metav1.CreateOptions{},
		)
	if err != nil {
		return nil, err
	}

	return sm, nil
}

func DeleteService(client kubernetes.Interface, namespace, name string) error {
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