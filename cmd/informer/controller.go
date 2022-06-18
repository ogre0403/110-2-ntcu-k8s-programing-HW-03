package main

import (
	"context"
	"fmt"
	"time"

	"os"
	"path"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	aa              *corev1.Service
	bb              *appv1.Deployment
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
	c.bb = createDeployment(c.clientSet)
	c.aa = createService(c.clientSet)
	fmt.Printf("Informer event: Job ADDED %s/%s\n", job.GetNamespace(), job.GetName())

}

func (c *DeploymentController) onUpdate(old, new interface{}) {
	job := old.(*appv1.Deployment)
	fmt.Printf("Informer event: Job UPDATED %s/%s\n", job.GetNamespace(), job.GetName())

	_, err := GetConfigMap(c.clientSet, namespace, "test-configmap")
	if err != nil && errors.IsNotFound(err) {
		cm, _ := CreateConfigMap(c.clientSet, namespace, "test-configmap")
		fmt.Printf("----Create ConfigMap when Job UPDATED Event %s/%s\n", cm.GetNamespace(), cm.GetName())
	}

}

func (c *DeploymentController) onDelete(obj interface{}) {
	job := obj.(*appv1.Deployment)
	fmt.Printf("Informer event: Job DELETED %s/%s\n", job.GetNamespace(), job.GetName())
	deleteDeployment(c.clientSet, c.bb)
	deleteService(c.clientSet, c.aa)

	if err := DeleteConfigMap(c.clientSet, namespace, "test-configmap"); err == nil {
		fmt.Printf("----Delete ConfigMap when Job DELETE Event %s/%s\n", namespace, "test-configmap")
	}
}

// NewConfigMapController creates a ConfigMapController
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

func createDeployment(client kubernetes.Interface) *appv1.Deployment {
	dm := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deploy",
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

var portnum int32 = 80

func createService(client kubernetes.Interface) *corev1.Service {
	sm := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "service",
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
					NodePort:   30001,
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



func GetConfigMap(client kubernetes.Interface, namespace, name string) (*corev1.ConfigMap, error) {
	cm, err := client.
		CoreV1().
		ConfigMaps(namespace).
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

func CreateConfigMap(client kubernetes.Interface, namespace, name string) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{Data: map[string]string{"foo": "bar"}}
	cm.Namespace = namespace
	cm.Name = name

	cm, err := client.
		CoreV1().
		ConfigMaps(namespace).
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

func DeleteConfigMap(client kubernetes.Interface, namespace, name string) error {
	err := client.
		CoreV1().
		ConfigMaps(namespace).
		Delete(
			context.Background(),
			name,
			metav1.DeleteOptions{},
		)

	return err
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
