package controller

import (
	"context"
	"log"
	"strings"
	"time"

	v1alpha1 "github.com/neha-Gupta1/step-controller/pkg/api/nehagupta.dev/v1alpha1"
	klientSet "github.com/neha-Gupta1/step-controller/pkg/client/clientset/versioned"
	informers "github.com/neha-Gupta1/step-controller/pkg/client/informers/externalversions/nehagupta.dev/v1alpha1"
	lister "github.com/neha-Gupta1/step-controller/pkg/client/listers/nehagupta.dev/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type Controller struct {
	Klient     klientSet.Interface
	StepSynced cache.InformerSynced
	StepLister lister.StepLister
	wq         workqueue.RateLimitingInterface
	client     kubernetes.Clientset
}

func NewController(klient klientSet.Interface, client kubernetes.Clientset, klusterInformer informers.StepInformer) *Controller {
	c := &Controller{
		Klient:     klient,
		StepSynced: klusterInformer.Informer().HasSynced,
		StepLister: klusterInformer.Lister(),
		wq:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultItemBasedRateLimiter(), "steps"),
		client:     client,
	}

	klusterInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.HandleAdd,
			DeleteFunc: c.HandleDelete,
		},
	)
	return c
}

func (c *Controller) HandleAdd(obj interface{}) {
	log.Println("New event captured")
	c.wq.Add(obj)
}

func (c *Controller) HandleDelete(obj interface{}) {
	log.Println("HandleDelete was called")
}

func (c *Controller) Run(ch chan struct{}) (err error) {

	if ok := cache.WaitForCacheSync(ch, c.StepSynced); !ok {
		log.Println("Cache not synced yet.")
	}

	go wait.Until(c.worker, time.Second, ch)

	<-ch
	return err
}

func (c *Controller) worker() {
	for c.processItems() {

	}
}

func (c *Controller) processItems() bool {
	item, shutDown := c.wq.Get()
	if shutDown {
		log.Println("got shutdown true")
		return false
	}

	defer c.wq.Forget(item)

	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		log.Println("Could not process item.Err :", err)
		return false
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Println("Could not process item.Err :", err)
		return false
	}
	step, err := c.StepLister.Steps(ns).Get(name)
	if err != nil {
		log.Println("Could not process item.Err :", err)
		return false
	}
	log.Printf("Steps we got : %+v", step.Spec)

	instance := step.DeepCopy()

	// Set At instance as the owner and controller
	owner := metav1.NewControllerRef(instance, v1alpha1.SchemeGroupVersion.WithKind("Step"))

	pod := newPodForCR(instance)
	pod.ObjectMeta.OwnerReferences = append(pod.ObjectMeta.OwnerReferences, *owner)

	// Try to see if the pod already exists and if not
	// (which we expect) then create a one-shot pod as per spec:
	found, err := c.client.CoreV1().Pods(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{})
	log.Println("Pod already present: ", found)
	log.Println("We are above if condition")
	if err != nil && errors.IsNotFound(err) {
		log.Println("We are inside if condition")
		found, err = c.client.CoreV1().Pods(pod.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
		if err != nil {
			log.Println("COuld not create pod: ", err)
			return false
		}
		klog.Infof("instance %s: pod launched: name=%s", key, pod.Name)
	}
	log.Println("Pod already present: ", found)

	return true
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *v1alpha1.Step) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "nginx-fromcr",
					Image:   "nginx:latest",
					Command: strings.Split(cr.Spec.Command, " "),
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}
