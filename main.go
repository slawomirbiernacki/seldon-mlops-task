package main

import (
	"context"
	"flag"
	"fmt"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"os"
	"seldon-mlops-task/seldonclient"
	"seldon-mlops-task/utils"
	"time"
)

//FIXME remove defaults
var namespaceFlag = flag.String("n", "test-aaa", "Namespace for your seldon deployment")
var deploymentFileFlag = flag.String("f", "test-resource.yaml", "Path to your deployment file")
var replicas = 2

func main() {
	flag.Parse()
	ctx := context.Background()

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

	//dereferencing after nil checks
	namespace := *namespaceFlag
	deploymentFile := *deploymentFileFlag

	fmt.Println("Deploying resource")
	deployment, err := createDeployment(ctx, namespace, deploymentFile)
	if err != nil {
		panic(err)
	}
	name := deployment.GetName()

	go seldonclient.WatchDeploymentEvents(ctx, deployment, namespace)

	fmt.Println("Waiting for deployment to become available...")
	err = seldonclient.WaitForDeploymentStatus(ctx, name, namespace, v1.StatusStateAvailable, time.Second, 100*time.Second)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Deployed, scalling to %d replicas \n", replicas)

	err = seldonclient.ScaleDeployment(ctx, name, namespace, replicas)
	if err != nil {
		panic(err)
	}

	fmt.Println("Waiting for pods...")
	err = seldonclient.WaitForScale(ctx, name, namespace, replicas, time.Second, 100*time.Second)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Scaled to %d replicas, deleting deployment\n", replicas)
	err = seldonclient.DeleteSeldonDeployment(ctx, name, namespace)
	if err != nil {
		panic(err)
	}

	fmt.Println("Waiting for deletion...")
	err = seldonclient.WaitUntilDeploymentDeleted(ctx, name, namespace, time.Second, 100*time.Second)
	if err != nil {
		panic(err)
	}

	fmt.Print("Deleted, program has finished!\n")
}

func createDeployment(ctx context.Context, namespace, deploymentFilePath string) (*v1.SeldonDeployment, error) {

	deployment := &v1.SeldonDeployment{}
	err := utils.ParseDeploymentFromFile(deploymentFilePath, deployment)
	if err != nil {
		return nil, err
	}

	deployment, err = seldonclient.CreateSeldonDeployment(ctx, deployment, namespace)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}
