package main

import (
	"context"
	"flag"
	"fmt"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	seldonclientset "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned"
	"github.com/slawomirbiernacki/seldon-mlops-task/seldondeployment"
	"github.com/slawomirbiernacki/seldon-mlops-task/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
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

	config, err := ctrl.GetConfig()
	if err != nil {
		panic(err)
	}
	seldonClientset, err := seldonclientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	manager := seldondeployment.NewManager(seldonClientset, namespace)

	fmt.Printf("Deploying resource %s to namespace %s\n", deploymentFile, namespace)
	deployment, err := createDeployment(ctx, manager, deploymentFile)
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

	processor := seldondeployment.NewEventProcessor(clientset, deployment, namespace)
	go watchDeploymentEvents(processor, &wg, quit)

	fmt.Printf("Waiting for deployment %s to become available...\n", name)
	err = manager.WaitUntilDeploymentStatus(ctx, name, v1.StatusStateAvailable, pollTimeout)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Deployed, scalling to %d replicas \n", replicas)

	err = manager.Scale(ctx, name, replicas)
	if err != nil {
		panic(err)
	}

	fmt.Println("Waiting for pods...")
	err = manager.WaitUntilDeploymentScaled(ctx, name, replicas, pollTimeout)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Scaled to %d replicas, deleting deployment\n", replicas)
	err = manager.DeleteDeployment(ctx, name)
	if err != nil {
		panic(err)
	}

	fmt.Println("Waiting for deletion...")
	err = manager.WaitUntilDeploymentDeleted(ctx, name, pollTimeout)
	if err != nil {
		panic(err)
	}

	fmt.Print("Deleted, program has finished!\n")
}

func createDeployment(ctx context.Context, manager *seldondeployment.Manager, deploymentFilePath string) (*v1.SeldonDeployment, error) {

	deployment := &v1.SeldonDeployment{}
	err := utils.ParseDeploymentFromFile(deploymentFilePath, deployment)
	if err != nil {
		return nil, err
	}

	deployment, err = manager.CreateDeployment(ctx, deployment)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func watchDeploymentEvents(processor *seldondeployment.EventProcessor, wg *sync.WaitGroup, quit chan bool) {
	defer wg.Done()

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
