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

### Building from source for local platform

        make

Build for other platform: (list of possible platforms https://golang.org/doc/install/source#environment)

        make PLATFORM=windows/amd64


