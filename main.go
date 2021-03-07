package main

import (
	"context"
	"flag"
	"fmt"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"seldon-mlops-task/operation"
	"seldon-mlops-task/utils"
	"time"
)

var namespace string
var deploymentFile string
var pollTimeout time.Duration
var replicas int

func init() {
	flag.StringVar(&namespace, "n", "default", "Namespace for your seldon deployment")
	flag.StringVar(&deploymentFile, "f", "test-resource.yaml", "Path to your deployment file")
	flag.DurationVar(&pollTimeout, "pt", 120*time.Second, "Poling timeout for any wait operations; eg waiting for deployment availability")
	flag.IntVar(&replicas, "r", 2, "Replica number to scale to during program operation")
}

func main() {
	flag.Parse()
	ctx := context.Background()

	fmt.Printf("Deploying resource to namespace %s\n", namespace)
	deployment, err := createDeployment(ctx, namespace, deploymentFile)
	if err != nil {
		panic(err)
	}
	name := deployment.GetName()

	go operation.WatchDeploymentEvents(ctx, deployment, namespace)

	fmt.Println("Waiting for deployment to become available...")
	err = operation.WaitUntilDeploymentStatus(ctx, name, namespace, v1.StatusStateAvailable, pollTimeout)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Deployed, scalling to %d replicas \n", replicas)

	err = operation.ScaleDeployment(ctx, name, namespace, replicas)
	if err != nil {
		panic(err)
	}

	fmt.Println("Waiting for pods...")
	err = operation.WaitUntilDeploymentScaled(ctx, name, namespace, replicas, pollTimeout)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Scaled to %d replicas, deleting deployment\n", replicas)
	err = operation.DeleteSeldonDeployment(ctx, name, namespace)
	if err != nil {
		panic(err)
	}

	fmt.Println("Waiting for deletion...")
	err = operation.WaitUntilDeploymentDeleted(ctx, name, namespace, pollTimeout)
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

	deployment, err = operation.CreateSeldonDeployment(ctx, deployment, namespace)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}
