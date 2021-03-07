# seldon-mlops-task


1. Prepare kubernetes cluster:

https://minikube.sigs.k8s.io/docs/start/

brew install kubernetes-helm


Install seldon-core https://github.com/SeldonIO/seldon-core

Install istio

https://istio.io/latest/docs/setup/getting-started/

install seldon core https://docs.seldon.io/projects/seldon-core/en/latest/workflow/install.html#pre-requisites

https://github.com/SeldonIO/seldon-core/tree/master/examples/auth


https://istio.io/latest/docs/setup/platform-setup/kind/

### Requirements
    Docker min v18.09
    Kubernetes cluster min v1.17.0 with istio and seldon core

### Building from source for local platform

        make

Build for other platform: (list of possible platforms https://golang.org/doc/install/source#environment)

        make PLATFORM=windows/amd64


### Usage
    Have kube config in .kube

    deploys to default namespace by default
    deploys test resource by default

use -h for help

    ./bin/app 
### Improvements

    switch to dynamic client to avoid dependencies
    build is long - dont know how to cache with mod and go sth 0
    use events for waiting

scaling arbitrary, no scalling orchestrator or components

event based instead of waiting

https://docs.seldon.io/projects/seldon-core/en/v1.1.0/examples/autoscaling_example.html

events - considered informer, considered filtering versions, chosen solution to mimic kubectl describe. 
ever growing list, prob not good for long running apps but here ok


could probably use background policy for removal