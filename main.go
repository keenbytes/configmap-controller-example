package main

import (
	"flag"
	"log"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

func main() {
	var kubeconfig string
	var annotation string

	// TODO: validation?

	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&annotation, "annotation", "", "annotation to look for")
	flag.Parse()

	var config *rest.Config
	var err error
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatal("error building kubeconfig")
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal("error getting incluster config")
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err, "error building kubernetes clientset")
	}

	configMapListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "configmaps", v1.NamespaceAll, fields.Everything())
	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]())

	indexer, informer := cache.NewInformerWithOptions(cache.InformerOptions{
		ListerWatcher: configMapListWatcher,
		ObjectType:    &v1.ConfigMap{},
		Handler: cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(old interface{}, new interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(new)
				if err == nil {
					queue.Add(key)
				}
			},
		},
		Indexers: cache.Indexers{},
	})

	ctl := NewController(queue, informer, indexer, annotation)

	ch := make(chan struct{})

	log.Printf("starting to watch annotation of '%s'", annotation)
	ctl.Run(ch)
}
