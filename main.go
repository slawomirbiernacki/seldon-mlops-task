package main

import (
	"context"
	"flag"
	"fmt"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"os"
	"seldon-mlops-task/seldonclient"
	"time"
)

var namespaceFlag = flag.String("n", "", "Namespace for your seldon deployment")
var deploymentFileFlag = flag.String("f", "", "Path to your deployment file")

func main() {
	flag.Parse()

	if namespaceFlag == nil || len(*namespaceFlag) == 0 {
		fmt.Println("Provide namespace")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if deploymentFileFlag == nil || len(*deploymentFileFlag) == 0 {
		fmt.Println("Provide deployment file")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Println("Starting deployment")
	ctx := context.Background()

	namespace := *namespaceFlag
	deploymentFile := *deploymentFileFlag

	listenForEvents("", namespace)

	deployment, err := createDeployment(ctx, namespace, deploymentFile)
	if err != nil {
		panic(err)
	}

	name := deployment.GetName()

	err = waitForDeployment(name, namespace, time.Second, 100*time.Second)
	if err != nil {
		panic(err)
	}

	// probably also check number of replicas and other stuff
	get, err := seldonclient.GetSeldonDeployment(context.TODO(), deployment.GetName(), namespace)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Deployed, status %s.\n", get.Status.State)

	// update like here : https://github.com/kubernetes/client-go/blob/master/examples/create-update-delete-deployment/main.go
	replicas := int32(2)
	get.Spec.Predictors[0].Replicas = nil
	get.Spec.Replicas = &replicas

	_, err = seldonclient.UpdateSeldonDeployment(context.TODO(), get, namespace)
	if err != nil {
		panic(err)
	}

	fmt.Print("Scaled up\n")

	get, err = seldonclient.GetSeldonDeployment(context.TODO(), deployment.GetName(), namespace)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Status after scalling up %s.\n", get.Status.State)

	//TODO wait until scaled up
	err = seldonclient.DeleteSeldonDeployment(context.TODO(), deployment.GetName(), namespace)
	if err != nil {
		panic(err)
	}
	fmt.Print("Deleted\n")
}

//FIXME filter by name
func listenForEvents(seldonDeployment, namespace string) {
	factory := seldonclient.NewInformerFactory(namespace)
	events := make(chan struct{})
	defer close(events)

	//FIXME cache sync? figure out how to start and stop correctly
	//factory.Start(wait.NeverStop)
	//factory.WaitForCacheSync(wait.NeverStop)

	informerr := factory.Machinelearning().V1().SeldonDeployments().Informer()

	go runSeldonCRDInformer(events, informerr, namespace)
}

func runSeldonCRDInformer(stopCh <-chan struct{}, s cache.SharedIndexInformer, namespace string) {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			d := obj.(*v1.SeldonDeployment)
			fmt.Printf("Added! %s.\n", d.GetName())

			// do what we want with the SeldonDeployment/event
		},
		DeleteFunc: func(obj interface{}) {
			d := obj.(*v1.SeldonDeployment)
			fmt.Printf("Deleted! %s.\n", d.GetName())
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			d := newObj.(*v1.SeldonDeployment)
			fmt.Printf("Updated! %s.\n", d.Status)
		},
	}
	s.AddEventHandler(handlers)
	s.Run(stopCh)
}

func createDeployment(ctx context.Context, namespace, deploymentFilePath string) (*v1.SeldonDeployment, error) {

	deployment, err := seldonclient.ParseDeploymentFromFile(deploymentFilePath)
	if err != nil {
		return nil, err
	}

	deployment, err = seldonclient.CreateSeldonDeployment(ctx, deployment, namespace)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Created deployment %q.\n", deployment.GetName())

	return deployment, nil
}

func waitForDeployment(name, namespace string, pollInterval, pollTimeout time.Duration) error {
	var reason string
	fmt.Print("Waiting\n")
	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		var err error
		deployment, err := seldonclient.GetSeldonDeployment(context.TODO(), name, namespace)
		if err != nil {
			return false, err
		}

		// When the deployment status and its underlying resources reach the desired state, we're done
		if v1.StatusStateAvailable == deployment.Status.State {
			//fmt.Print("\n")
			return true, nil
		}

		reason = fmt.Sprintf("deployment status: %#v", deployment.Status)

		//fmt.Print(".")
		return false, nil
	})

	if err == wait.ErrWaitTimeout {
		err = fmt.Errorf("%s", reason)
	}
	if err != nil {
		return fmt.Errorf("error waiting for deployment %q status to match expectation: %v", name, err)
	}
	return nil
}
