# configmap-controller-example
Quick sample of Kubernetes controller that monitors ConfigMap resources with a specified annotation 
and calls URL when change happens.

This code is based on the [sample-controller](sample-controller) repository available from official
Kubernetes repository, however, it does not have any Custom Resource Definition.

## Build
Run `go build -o myctlr .` to build the binary.

## Run
Modify the following command to run the controller.

    ./myctlr --kubeconfig=$PATH_TO_KUBECONFIG --annotation=myctlr/onchange-url

where `$PATH_TO_KUBECONFIG` points to kubeconfig, eg. `/home/mikolaj/.kube/config` and `annotation` is
name of the annotation that contains URL that should be called.

### kubeconfig
The `--kubeconfig` argument is optional and if not passed, the controller will assume it is running
inside of the cluster.

### annotation
See below example of ConfigMap using annotation of `myctlr/onchange-url` that would be monitored.

    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: sample-config
      namespace: test-watcher
      annotations:
        myctlr/onchange-url: 'https://neverssl.com'
    data:
      key1: val1
      key2: val2

## Modify
Check the `processQueue` function in `controller.go` file and replace the code with the desired
functionality.  ConfigMap object is fetched so its contents can be used.  Also, if there should be
more actions happening when change occurs, second annotation can be added.  For example, you might want
to have `myctlr/onchange-restart` with a name of Deployment that should be restarted, and pass only
the `myctlr` in `annotation` argument.

## Appendix
Check `sample-Dockerfile` for building docker image and `sample-deployment.yaml` for Deployment in
Kubernetes.
