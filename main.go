package main

import (
	"context"
	"flag"
	"fmt"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	corev1 "k8s.io/api/core/v1"
	"seldon-mlops-task/seldondeployment"
	"seldon-mlops-task/utils"
	"sync"
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

	// Here's some setup for parallel event processing.
	// I use wait group and quit channel to allow event processor loop to exit gracefully. This ensures all events are processed before the program finishes.
	// Quit channel is passed to the goroutine and internal loop exits when a signal is sent to it.
	// The goroutine then notifies the wait group on which main routine is waiting at the end of the program.
	quit := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(1)
	defer func() {
		quit <- true
		wg.Wait()
	}()
	go watchDeploymentEvents(deployment, namespace, &wg, quit)

	fmt.Println("Waiting for deployment to become available...")
	err = seldondeployment.WaitUntilDeploymentStatus(ctx, name, namespace, v1.StatusStateAvailable, pollTimeout)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Deployed, scalling to %d replicas \n", replicas)

	err = seldondeployment.Scale(ctx, name, namespace, replicas)
	if err != nil {
		panic(err)
	}

	fmt.Println("Waiting for pods...")
	err = seldondeployment.WaitUntilDeploymentScaled(ctx, name, namespace, replicas, pollTimeout)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Scaled to %d replicas, deleting deployment\n", replicas)
	err = seldondeployment.DeleteDeployment(ctx, name, namespace)
	if err != nil {
		panic(err)
	}

	fmt.Println("Waiting for deletion...")
	err = seldondeployment.WaitUntilDeploymentDeleted(ctx, name, namespace, pollTimeout)
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

	deployment, err = seldondeployment.CreateDeployment(ctx, deployment, namespace)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func watchDeploymentEvents(deployment *v1.SeldonDeployment, namespace string, wg *sync.WaitGroup, quit chan bool) {
	defer wg.Done()
	processor := seldondeployment.NewEventProcessor(deployment, namespace)
	for {
		select {
		case <-quit:
			return
		default:
			err := processor.ProcessNew(printEvent)
			if err != nil {
				panic(err)
			}
			time.Sleep(time.Second)
		}
	}
}

func printEvent(event corev1.Event) error {
	fmt.Printf("Event: UUID: %s, Type: %s, FROM: %s, Reason: %s, message:%s, \n", event.UID, event.Type, event.Source.Component, event.Reason, event.Message)
	return nil
}
