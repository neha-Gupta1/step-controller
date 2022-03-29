package main

import (
	"flag"
	"log"
	"path/filepath"
	"time"

	klient "github.com/neha-Gupta1/step-controller/pkg/client/clientset/versioned"
	informerFactory "github.com/neha-Gupta1/step-controller/pkg/client/informers/externalversions"
	controller "github.com/neha-Gupta1/step-controller/pkg/controller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {

	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("Building config from flags failed, %s, trying to build inclusterconfig", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Printf("error %s building inclusterconfig", err.Error())
			return
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("getting Klient set %s client %v\n", err.Error(), client)
	}
	clientValue := *client
	klientSet, err := klient.NewForConfig(config)
	if err != nil {
		log.Printf("getting Klient set %s client %v\n", err.Error(), klientSet)
	}

	informerFactory := informerFactory.NewSharedInformerFactory(klientSet, 20*time.Minute)
	ch := make(chan struct{})
	cntlr := controller.NewController(klientSet, clientValue, informerFactory.Nehagupta().V1alpha1().Steps())

	informerFactory.Start(ch)
	if err := cntlr.Run(ch); err != nil {
		log.Println("Error found: ", err)
	}

}
