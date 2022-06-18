package main

import (
	"context"
	"fmt"
	"time"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	
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
	svc              *corev1.Service
	dep              *appv1.Deployment
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
	j := obj.(*appv1.Deployment)
	if (j.GetLabels()["ntcu-k8s"] != "hw3") {
		return
	}
	c.dep = createDeployment(c.clientSet)
	c.svc = createService(c.clientSet)
	fmt.Printf("Informer event: Deployment ADDED %s/%s\n", j.GetNamespace(), j.GetName())
}

func (c *DeploymentController) onUpdate(old, new interface{}) {
	j := old.(*appv1.Deployment)
	if (j.GetLabels()["ntcu-k8s"] != "hw3") {
		return
	}
	fmt.Printf("Informer event: Deploymenttttttt UPDATED %s/%s\n", j.GetNamespace(), j.GetName())

}

func (c *DeploymentController) onDelete(obj interface{}) {
	j := obj.(*appv1.Deployment)
	if (j.GetLabels()["ntcu-k8s"] != "hw3") {
		return
	}
	fmt.Printf("Informer event: Deployment DELETED %s/%s\n", j.GetNamespace(), j.GetName())
	deleteDeployment(c.clientSet, c.dep)
	deleteService(c.clientSet, c.svc)
	
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

var portnum int32 = 80

func int32Ptr(i int32) *int32 { return &i }

func createService(client kubernetes.Interface) *corev1.Service {
	sm := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "imservice",
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

func createDeployment(client kubernetes.Interface) *appv1.Deployment {
	dm := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "imdeploy",
			Labels: map[string]string{
				"ntcu-k8s": "hw3",
			},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"ntcu-k8s": "hw3",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"ntcu-k8s": "hw3",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx-container",
							Image: "nginx:1.14.2",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}
	dm.Namespace = namespace

	dm, err := client.
		AppsV1().
		Deployments(namespace).
		Create(
			context.Background(),
			dm,
			metav1.CreateOptions{},
		)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Created Deployment %s/%s\n", dm.GetNamespace(), dm.GetName())
	return dm
}

func deleteDeployment(client kubernetes.Interface, dm *appv1.Deployment) {
	err := client.
		AppsV1().
		Deployments(dm.GetNamespace()).
		Delete(
			context.Background(),
			dm.GetName(),
			metav1.DeleteOptions{},
		)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Deleted Deployment %s/%s\n", dm.GetNamespace(), dm.GetName())
}
