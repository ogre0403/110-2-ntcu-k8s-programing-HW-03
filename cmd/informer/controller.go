
package main
import (
    "fmt"
    "time"
    appv1 "k8s.io/api/apps/v1"
    //batchv1 "k8s.io/api/batch/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/client-go/informers"
    batchinformer "k8s.io/client-go/informers/apps/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/cache"
    //"k8s.io/apimachinery/pkg/util/wait"
)
type DeploymentController struct {
    informerFactory informers.SharedInformerFactory
    informer        batchinformer.DeploymentInformer
    clientSet       *kubernetes.Clientset
    svc             *corev1.Service
    deployment      *appv1.Deployment
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
    job := obj.(*appv1.Deployment)
    if job.GetLabels()["ntcu-k8s"] == "hw3" {
    fmt.Printf("Informer event: Job ADDED %s/%s\n", job.GetNamespace(), job.GetName())
}
}
func (c *DeploymentController) onUpdate(old, new interface{}) {
    job := old.(*appv1.Deployment)
    if job.GetLabels()["ntcu-k8s"] == "hw3" {
    fmt.Printf("Informer event: Job UPDATED %s/%s\n", job.GetNamespace(), job.GetName())
        c.svc = createService(c.clientSet)
        c.deployment = createDeployment(c.clientSet)
        fmt.Printf("----Create Service when Job UPDATED Event %s/%s\n", c.svc.GetNamespace(), c.svc.GetName())
    }
}
func (c *DeploymentController) onDelete(obj interface{}) {
    job := obj.(*appv1.Deployment)
    if job.GetLabels()["ntcu-k8s"] == "hw3" {
    fmt.Printf("Informer event: Job DELETED %s/%s\n", job.GetNamespace(), job.GetName())
        deleteService(c.clientSet, c.svc)
        deleteDeployment(c.clientSet, c.deployment)
        fmt.Printf("----Delete Service when Job DELETE Event %s/%s\n", namespace, "test-service")
    }
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
    //factory.Start(wait.NeverStop)
    //factory.WaitForCacheSync(wait.NeverStop)
    //informer.Lister().Deployments("default").Get("default")
    return c
}
