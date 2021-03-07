package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	informer "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/informers/externalversions"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
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

	//listenForEvents("", namespace)

	deployment, err := createDeployment(ctx, namespace, deploymentFile)
	if err != nil {
		panic(err)
	}

	go watchDeployment(ctx, deployment, namespace)

	//clientset, err := kubernetes.NewForConfig(seldonclient.Config)
	//if err != nil {
	//	panic(err)
	//}

	//dep, err := seldonclient.GetSeldonDeployment(ctx, name, namespace)

	//scheme :=runtime.NewScheme()
	//v1.AddToScheme(scheme)
	//events, err := clientset.CoreV1().Events(namespace).Search(scheme, deployment)
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Printf("Events count: %d\n", len(events.Items))

	name := deployment.GetName()

	err = seldonclient.WaitForDeploymentStatus(ctx, name, namespace, v1.StatusStateAvailable, time.Second, 100*time.Second)
	if err != nil {
		panic(err)
	}

	//get, err := seldonclient.GetSeldonDeployment(ctx, name, namespace)
	//if err != nil {
	//	panic(err)
	//}
	//events, err = clientset.CoreV1().Events(namespace).Search(scheme, get)
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Printf("Events count2: %d\n", len(events.Items))

	replicas := 2

	err = seldonclient.ScaleDeployment(ctx, name, namespace, replicas)
	if err != nil {
		panic(err)
	}

	err = seldonclient.WaitForScale(ctx, name, namespace, replicas, time.Second, 100*time.Second)
	if err != nil {
		panic(err)
	}

	fmt.Print("Remove!\n")
	err = seldonclient.DeleteSeldonDeployment(ctx, name, namespace)
	if err != nil {
		panic(err)
	}

	//TODO wait for delete!!!!!
	//prompt()
	//_, err = seldonclient.GetSeldonDeployment(ctx, name, namespace)
	//if err != nil {
	//	panic(err)
	//}
	fmt.Print("Waiting for removal!\n")
	seldonclient.WaitUntilDeploymentDeleted(ctx, name, namespace, time.Second, 100*time.Second)

	fmt.Print("Finished!\n")
}

func watchDeployment(ctx context.Context, deployment *v1.SeldonDeployment, namespace string) {
	clientset, err := kubernetes.NewForConfig(seldonclient.Config)
	if err != nil {
		panic(err)
	}

	scheme := runtime.NewScheme()
	v1.AddToScheme(scheme)
	//var lastEventsVersion = ""

	seen := make(map[types.UID]bool)

	for i := 0; i < 160; i++ { //implement proper loop
		events, err := clientset.CoreV1().Events(namespace).Search(scheme, deployment)
		if err != nil {
			panic(err)
		}

		for _, event := range events.Items {

			if !seen[event.UID] {
				fmt.Printf("Event: UUID: %s, Version:%s,  Type: %s, FROM: %s, Reason: %s, message:%s, \n", event.UID, event.ResourceVersion, event.Type, event.Source.Component, event.Reason, event.Message)
				seen[event.UID] = true
			}

			//if len(lastEventsVersion) == 0{
			//	fmt.Printf("Event: UUID: %s, Version:%s,  Type: %s, Reason: %s, message:%s, \n", event.UID, event.ResourceVersion, event.Type, event.Reason,event.Message)
			//}else{
			//	if event.ResourceVersion > lastEventsVersion{
			//		fmt.Printf("Event2: UUID: %s, Version:%s,  Type: %s, Reason: %s, message:%s, \n", event.UID, event.ResourceVersion, event.Type, event.Reason,event.Message)
			//	}
			//}

		}
		//lastEventsVersion = events.ResourceVersion
		//fmt.Printf("lastEventsVersion updated to: %s \n", lastEventsVersion)

		time.Sleep(2 * time.Second)
	}
}

//FIXME filter by name!
func listenForEvents(seldonDeployment, namespace string) {
	factory := seldonclient.NewInformerFactory(namespace)
	events := make(chan struct{})
	//FIXME
	//defer close(events)

	//FIXME cache sync? figure out how to start and stop correctly

	informerr := factory.Machinelearning().V1().SeldonDeployments().Informer()
	runSeldonCRDInformer(events, informerr, namespace)
	factoryStart(factory) // coroutine?
}

func factoryStart(factory informer.SharedInformerFactory) {
	factory.WaitForCacheSync(wait.NeverStop)
	factory.Start(wait.NeverStop)
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
	//s.Run(stopCh)
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
