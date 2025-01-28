package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	queue      workqueue.TypedRateLimitingInterface[string]
	informer   cache.Controller
	indexer    cache.Store
	annotation string
}

func NewController(queue workqueue.TypedRateLimitingInterface[string], informer cache.Controller, indexer cache.Store, annotation string) *Controller {
	return &Controller{
		informer:   informer,
		indexer:    indexer,
		queue:      queue,
		annotation: annotation,
	}
}

func (c *Controller) Run(ch chan struct{}) {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		c.informer.Run(ch)
		wg.Done()
	}()

	defer c.queue.ShutDown()
	defer close(ch)

	for c.processQueue() {
	}

	wg.Wait()
}

func (c *Controller) processQueue() bool {
	key, shutdown := c.queue.Get()
	if shutdown {
		log.Print("queue shutdown")
		return false
	}
	defer c.queue.Done(key)

	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		log.Print("error getting key from indexer")
		return true
	}
	if !exists {
		log.Print("key does not exist anymore")
		return true
	}

	configMap := obj.(*v1.ConfigMap).DeepCopy()
	url, ok := configMap.Annotations[c.annotation]
	if !ok {
		return true
	}
	log.Printf("found configmap with annotation %s with value of '%s'", c.annotation, url)

	// TODO: validation of the value

	// Add your desired action here
	err = c.requestURL(url)
	if err != nil {
		log.Printf("error with requesting url: %s", err.Error())
		return true
	}
	log.Printf("url %s successfully called", url)

	return true
}

func (c *Controller) requestURL(url string) error {
	req, err := http.NewRequest("GET", url, strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("error creating get request: %w", err)
	}
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("error making get request: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("response status code is different than 200")
	}

	return nil
}
