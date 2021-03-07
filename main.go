package main

import (
	"context"
	"flag"
	"fmt"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
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

	deployment, err := createDeployment(ctx, namespace, deploymentFile)
	if err != nil {
		panic(err)
	}

	go watchDeployment(ctx, deployment, namespace)

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

	err = seldonclient.WaitForScale(ctx, name, namespace, replicas, time.Second, 100*time.Second)
	if err != nil {
		panic(err)
	}

	fmt.Print("Remove!\n")
	err = seldonclient.DeleteSeldonDeployment(ctx, name, namespace)
	if err != nil {
		panic(err)
	}

	fmt.Print("Waiting for removal!\n")
	err = seldonclient.WaitUntilDeploymentDeleted(ctx, name, namespace, time.Second, 100*time.Second)
	if err != nil {
		panic(err)
	}
	fmt.Print("Finished!\n")
}

func watchDeployment(ctx context.Context, deployment *v1.SeldonDeployment, namespace string) {
	clientset, err := kubernetes.NewForConfig(seldonclient.Config)
	if err != nil {
		panic(err)
	}

	scheme := runtime.NewScheme()
	err = v1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
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
