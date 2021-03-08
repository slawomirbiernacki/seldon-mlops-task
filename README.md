## General information

Here's my simple program that deploys seldon custom resource to kubernetes cluster.

### Assumptions and general thoughts

* I used `clientset` from seldon-core operator module to interact with the cluster. 
  It's a quite heavy dependency so given we don't do much seldon domain specific operations a generic dynamic client 
  could have been used instead. I did not optimize for binary size.
* For watching events I used generic `clientset.CoreV1().Events()` interface, polling it in a simple loop to mimic a bit `kubectl describe`as it has human-readable descriptions. Arbitrary decision - each event is printed only once, ignoring resource versions.
* I have also explored using Seldon specific `SharedIndexInformer` but it was difficult to produce meaningful descriptions for update events. 
  However, if event watching functionality were to be used for any application logic, `SharedIndexInformer` would have been preferred.
* I used polling when waiting for the deployment to become available or reach target scale. 
  I can imagine alternative design where program operations observe events and react to them, eg issuing delete once replicas count has reached target number.
* In Seldon Core documentation I can see that deployments can specify replicas on [multiple levels](https://docs.seldon.io/projects/seldon-core/en/v1.1.0/graph/scaling.html).
  My scaling logic is quite arbitrary - it overrides any given setting with top level `.spec.replicas`. (Except `.spec.predictors[].svcOrchSpec.replicas` which I ignore).
* I use `DeletePropagationForeground`policy when deleting to wait until deployment is fully deleted. 
  However, event interface I use doesn't produce any events for deletions so probably could be changes to a background one.
* I included an example test for scaling logic. For production code more tests would be needed.

### Requirements

* Kubernetes cluster >= `v1.17.0` with Seldon Core installed
* Configured authentication to the cluster through kubectl config (see kubectl [documentation](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/#verify-kubectl-configuration) for details)

### How to use it

* Binaries for popular platforms are available in the release, go [github releases](https://github.com/slawomirbiernacki/seldon-mlops-task/releases/tag/v1.0.0) to download.
To run the program with defaults using included test deployment file, run:
  
        For mac:
        ./app-darwin-amd64

        For linux
        ./app-linux-amd64
  
That will deploy `test-resource.yaml` to `default` namespace

If you use a different platform, see [Build from sources](#Build from sources)

Run `./app-{your platform} -h` to see a list of flags that can be used to configure the program

| flag | function                                                                       | default value      |
|------|--------------------------------------------------------------------------------|--------------------|
| -f   | Path to your deployment file                                                   | test-resource.yaml |
| -n   | Namespace for your seldon deployment                                           | default            |
| -r   | Replica number to scale to during program operation                            | 2                  |
| -pt  | Poling timeout for any wait operations; eg waiting for deployment availability | 120s               |

There's also `-kubeconfig` which should be respected when looking up kubernetes connection configuration but haven't tested it.

Run with your deployment file in your namespace:

        ./app-{your platform} -f your-file.yaml -n test-namespace

### Build from sources

Building can be done in two ways - using local Go installation or docker. 

#### Build using local Go installation

* Required go installed in a version >= `1.14`
* Run `make build-dev` to compile for local platform
* Compiled binary will be available in `/bin`

#### Build using docker

* Required docker installation in a version >= `v18.09`
* Run `make build` to compile for local platform
  * Alternatively run `make build PLATFORM=linux/amd64` to specify target platform
  * See [list of available platforms](https://golang.org/doc/install/source#environment)
* Compiled binary will be available in `/bin`
* Go make yourself some tea, it takes a while.