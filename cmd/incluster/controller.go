package main

import (
	"context"
	"fmt"
	"time"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/informers"
	batchinformer "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type DeploymentController struct {
	informerFactory informers.SharedInformerFactory
	informer        batchinformer.DeploymentInformer
	clientSet       *kubernetes.Clientset
	aa              *corev1.Service
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
	dd := obj.(*appv1.Deployment)
	if !(dd.GetLabels()["ntcu-k8s"] == "hw3") {
		return
	}
	c.aa = createService(c.clientSet, dd)
	fmt.Printf("Informer event: Deployment ADDED %s/%s\n", dd.GetNamespace(), dd.GetName())
}

func (c *DeploymentController) onUpdate(old, new interface{}) {
	dd := old.(*appv1.Deployment)
	fmt.Printf("Informer event: Deploymenttttttt UPDATED %s/%s\n", dd.GetNamespace(), dd.GetName())

	// _, err := util.GetConfigMap(c.clientSet, namespace, "test-configmap")
	// if err != nil && errors.IsNotFound(err) {
	// 	cm, _ := util.CreateConfigMap(c.clientSet, namespace, "test-configmap")
	// 	fmt.Printf("----Create ConfigMap when Job UPDATED Event %s/%s\n", cm.GetNamespace(), cm.GetName())
	// }

}

func (c *DeploymentController) onDelete(obj interface{}) {
	dd := obj.(*appv1.Deployment)
	fmt.Printf("Informer event: Deployment DELETED %s/%s\n", dd.GetNamespace(), dd.GetName())
	deleteService(c.clientSet, c.aa)
	// if err := util.DeleteConfigMap(c.clientSet, namespace, "test-configmap"); err == nil {
	// 	fmt.Printf("----Delete ConfigMap when Job DELETE Event %s/%s\n", namespace, "test-configmap")
	// }
}

// NewConfigMapController creates a ConfigMapController
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

var portnum int32 = 80

func int32Ptr(i int32) *int32 { return &i }

func createService(client kubernetes.Interface, dd *appv1.Deployment) *corev1.Service {

	sm := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "banana",
			Labels: map[string]string{
				"ntcu-k8s": "hw3",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: dd.Spec.Selector.MatchLabels,
			Type:     corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.IntOrString{IntVal: portnum},
					NodePort:   30010,
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	sm.Namespace = namespace
	sm, err := client.
		CoreV1().
		Services(namespace).Create(
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

func deleteService(client kubernetes.Interface, sm *corev1.Service) {
	err := client.
		CoreV1().
		Services(sm.GetNamespace()).
		Delete(
			context.Background(),
			sm.GetName(),
			metav1.DeleteOptions{},
		)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Deleted Service %s/%s\n", sm.GetNamespace(), sm.GetName())
}
