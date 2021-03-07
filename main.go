package main

import (
	"context"
	"flag"
	"fmt"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"seldon-mlops-task/seldonclient"
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
