package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"os"
	"seldon-mlops-task/seldonclient"
	"time"
)

//FIXME remove defaults
var namespaceFlag = flag.String("n", "test-aaa", "Namespace for your seldon deployment")
var deploymentFileFlag = flag.String("f", "test-resource.yaml", "Path to your deployment file")

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

	err = seldonclient.WaitForDeploymentStatus(ctx, name, namespace, v1.StatusStateAvailable, time.Second, 100*time.Second)
	if err != nil {
		panic(err)
	}

	replicas := 2

	err = seldonclient.ScaleDeployment(ctx, name, namespace, replicas)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(seldonclient.Config)
	dep, err := seldonclient.GetSeldonDeployment(ctx, name, namespace)
	events, err := clientset.CoreV1().Events(namespace).Search(runtime.NewScheme(), dep)

	print(events)

	err = seldonclient.WaitForScale(ctx, name, namespace, replicas, time.Second, 100*time.Second)
	if err != nil {
		panic(err)
	}

	err = seldonclient.DeleteSeldonDeployment(ctx, name, namespace)
	if err != nil {
		panic(err)
	}
	fmt.Print("Finished!\n")
}

//FIXME filter by name!
func listenForEvents(seldonDeployment, namespace string) {
	factory := seldonclient.NewInformerFactory(namespace)
	events := make(chan struct{})
	//FIXME
	//defer close(events)

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
			fmt.Printf("Updated! %v\n", d.Status.Description)
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

func prompt() {
	fmt.Printf("-> Press Return key to continue.\n")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
