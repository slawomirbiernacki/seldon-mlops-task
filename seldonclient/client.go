package seldonclient

import (
	"context"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	seldonclientset "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned"
	informer "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"

	// imported all authentication handlers
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

//TODO possibly wrap whole seldon in a struct?
var seldonClientset *seldonclientset.Clientset

//FIXME remove
var Config *rest.Config

func init() {
	var err error
	seldonClientset, err = getSeldonClientSet()

	if err != nil {
		panic(err)
	}
}

func GetSeldonDeployment(ctx context.Context, name string, namespace string) (*v1.SeldonDeployment, error) {
	return seldonClientset.MachinelearningV1().SeldonDeployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

func DeleteSeldonDeployment(ctx context.Context, name string, namespace string) (err error) {
	return seldonClientset.MachinelearningV1().SeldonDeployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func CreateSeldonDeployment(ctx context.Context, deployment *v1.SeldonDeployment, namespace string) (*v1.SeldonDeployment, error) {
	return seldonClientset.MachinelearningV1().SeldonDeployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
}

func UpdateSeldonDeployment(ctx context.Context, deployment *v1.SeldonDeployment, namespace string) (*v1.SeldonDeployment, error) {
	return seldonClientset.MachinelearningV1().SeldonDeployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
}

//FIXME filter only to current deployment, what with the resyncing time?
func NewInformerFactory(namespace string) informer.SharedInformerFactory {
	return informer.NewSharedInformerFactoryWithOptions(seldonClientset, time.Second*30, informer.WithNamespace(namespace))
}

func getSeldonClientSet() (*seldonclientset.Clientset, error) {

	config := ctrl.GetConfigOrDie()

	kubeClientset, err := seldonclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	Config = config

	return kubeClientset, nil
}
