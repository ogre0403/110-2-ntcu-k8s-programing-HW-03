package main

import (
	"context"
	"fmt"
	"time"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	deploymentinformer "k8s.io/client-go/informers/apps/v1"
)

type DeploymentController struct {
	informerFactory informers.SharedInformerFactory
	informer        deploymentinformer.DeploymentInformer
	clientSet       *kubernetes.Clientset
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
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
	deployment := obj.(*appv1.Deployment)

	if !(deployment.GetLabels()["ntcu-k8s"] == "hw3") {
		return
	}

	CreateServices(c.clientSet, namespace, "nginx-service", deployment)
	fmt.Printf("Informer event: Deployment Added %s/%s\n", deployment.GetNamespace(), deployment.GetName())
}

func (c *DeploymentController) onUpdate(old, new interface{}) {
	deployment := old.(*appv1.Deployment)

	if !(deployment.GetLabels()["ntcu-k8s"] == "hw3") {
		return
	}
	fmt.Printf("Informer event: Deployment Update %s/%s\n", deployment.GetNamespace(), deployment.GetName())

	_, err := GetService(c.clientSet, namespace, "nginx-service")
	if err != nil && errors.IsNotFound(err) {
		cm, _ := CreateServices(c.clientSet, namespace, "nginx-service", deployment)
		fmt.Printf("---Create Service when Deployment Update Event %s/%s\n", cm.GetNamespace(), cm.GetName())
	}
}

func (c *DeploymentController) onDelete(obj interface{}) {
	deployment := obj.(*appv1.Deployment)
	if !(deployment.GetLabels()["ntcu-k8s"] == "hw3") {
		return
	}
	fmt.Printf("Informer event: Deployment Delete %s/%s\n", deployment.GetNamespace(), deployment.GetName())

	if err := DeleteService(c.clientSet, namespace, "nginx-service"); err == nil {
		fmt.Printf("---Delete Service when Deployment Delete Event %s/%s\n", namespace, "nginx-service")
	}
}

func NewDeploymentController(client *kubernetes.Clientset) *DeploymentController {
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

func GetService(client kubernetes.Interface, namespace, name string) (*corev1.Service, error) {
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

var portnum int32 = 80

func CreateServices(client kubernetes.Interface, namespace, name string, deployment *appv1.Deployment) (*corev1.Service, error) {
	sm := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"ntcu-k8s": "hw3",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: deployment.Spec.Selector.MatchLabels,
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