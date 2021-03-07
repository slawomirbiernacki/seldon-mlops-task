package seldonclient

import (
	"context"
	"fmt"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	seldonclientset "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned"
	clientv1 "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned/typed/machinelearning.seldon.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"

	// import all authentication handlers
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

//TODO possibly wrap whole seldon in a struct?
var seldonClientset *seldonclientset.Clientset
var clientset *kubernetes.Clientset
var deletePolicy = metav1.DeletePropagationForeground

func init() {
	config := ctrl.GetConfigOrDie()

	var err error
	seldonClientset, err = seldonclientset.NewForConfig(config)

	if err != nil {
		panic(err)
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
}

func GetSeldonDeployment(ctx context.Context, name string, namespace string) (*v1.SeldonDeployment, error) {
	return deploymentInterface(namespace).Get(ctx, name, metav1.GetOptions{})
}

func DeleteSeldonDeployment(ctx context.Context, name string, namespace string) (err error) {
	return deploymentInterface(namespace).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

func CreateSeldonDeployment(ctx context.Context, deployment *v1.SeldonDeployment, namespace string) (*v1.SeldonDeployment, error) {
	return deploymentInterface(namespace).Create(ctx, deployment, metav1.CreateOptions{})
}

func UpdateSeldonDeployment(ctx context.Context, deployment *v1.SeldonDeployment, namespace string) (*v1.SeldonDeployment, error) {
	return deploymentInterface(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
}

func deploymentInterface(namespace string) clientv1.SeldonDeploymentInterface {
	return seldonClientset.MachinelearningV1().SeldonDeployments(namespace)
}

func WatchDeploymentEvents(ctx context.Context, deployment *v1.SeldonDeployment, namespace string) {
	scheme := runtime.NewScheme()
	err := v1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

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
		}

		time.Sleep(2 * time.Second)
	}
}
