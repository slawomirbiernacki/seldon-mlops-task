## General information

Here's my simple program that deploys seldon custom resource to kubernetes cluster and manipulates it.

## Assumptions and general thoughts

* I made a number of arbitrary decisions, explained below. Would be great to hear feedback on that given my limited experience with kubernetes.
* I used `clientset` from seldon-core operator module to interact with the cluster. 
  The program doesn't do much seldon domain specific operations, so a generic dynamic client 
  could have been used instead.
* I didn't optimize for binary size.
* For watching events I used generic `clientset.CoreV1().Events()` interface, polling it in a simple loop to mimic a bit `kubectl describe`as it has human-readable descriptions. 
  Arbitrary decision - each event is printed only once, ignoring resource versions.
* I have also explored using Seldon specific `SharedIndexInformer` for handling events but it was difficult to produce readable descriptions for update events. 
  However, if event watching functionality were to be used for any application logic, `SharedIndexInformer` would have been preferred.
* I used polling when waiting for the deployment to become available or reach target scale. 
  I can imagine alternative design where the program observes events and react to them, eg issuing delete once replicas count has reached target number.
* In Seldon Core documentation I can see that deployments can specify replicas on [multiple levels](https://docs.seldon.io/projects/seldon-core/en/v1.1.0/graph/scaling.html).
  My scaling logic is simplified - it overrides any given setting with top level `.spec.replicas`. (Except `.spec.predictors[].svcOrchSpec.replicas` which I ignore).
* I use `DeletePropagationForeground`policy when deleting to wait until deployment is fully deleted. 
  However, the event interface I used doesn't produce any events for deletions - the policy probably could be changed to a background one.
* I included an example test for scaling logic. For production code more tests would be needed.

## Requirements

* Kubernetes cluster >= `v1.17.0` with Seldon Core installed
* Configured authentication to the cluster, eg through kubeconfig (see kubectl [documentation](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) for details).
  Program uses `~/.kube/config` by default.

## How to use it

There are 2 ways to run the program:
1. Install it using local Go installation
2. Build a binary using Docker

### Install it locally using Go

* Requires Go (>=`1.14`)
* Clone the repository and `cd` inside it.
  * Normally one can install go modules without cloning them manually. However, I have a `replace` directive present in go module 
    definition, causing `go install github.com/slawomirbiernacki/seldon-mlops-task` to fail. Is there a way to fix that issue? Perhaps, but I'm not that familiar with go modules (yet!).
* Run `go install .`.
* That will install the binary in your `$GOPATH/bin`.
* If you have above on your '$PATH' simply run it:
        
        seldon-mlops-task

* Otherwise provide full path:

        $GOPATH/bin/seldon-mlops-task

By default, the program tries to deploy provided `test-resource.yaml` to `default` namespace, the resource file should be in the directory from where you run the program. 
Otherwise, use a flag to point to it. Use `seldon-mlops-task -h` to see available flags.

| flag        | function                                                                       | default value      |
|-------------|--------------------------------------------------------------------------------|--------------------|
| -f          | Path to your deployment file                                                   | test-resource.yaml |
| -n          | Namespace for your seldon deployment                                           | default            |
| -r          | Replica number to scale to during program operation                            | 2                  |
| -pt         | Poling timeout for any wait operations; eg waiting for deployment availability | 120s               |
| -kubeconfig | Path to kubeconfig                                                             | ~/.kube.config     |

Example running your deployment file in your namespace:

        ./seldon-mlops-task -f your-file.yaml -n your-namespace

### Build a binary using Docker

If you don't have Go installed, you can build a binary using Docker.

* Requires `GNU Make`
* Requires Docker ( >=`v18.09`)
* Clone the repository and `cd` inside it.
* Run `make build` to compile for local platform
  * Alternatively run `make build PLATFORM=linux/amd64` to specify target platform
  * See [list of available platforms](https://golang.org/doc/install/source#environment)
* Compiled binary will be available in `/bin`
* Might take a while on the first run, consider having a biscuit.