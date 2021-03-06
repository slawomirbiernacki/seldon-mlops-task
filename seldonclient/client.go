package seldonclient

import (
	"context"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	seldonclientset "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned"
	informer "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"

	// imported all authentication handlers
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

//TODO possibly wrap whole seldon in a struct?
var clientset *seldonclientset.Clientset

func init() {
	var err error
	clientset, err = getSeldonClientSet()
	if err != nil {
		panic(err)
	}
}

func GetSeldonDeployment(ctx context.Context, name string, namespace string) (*v1.SeldonDeployment, error) {
	return clientset.MachinelearningV1().SeldonDeployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

func DeleteSeldonDeployment(ctx context.Context, name string, namespace string) (err error) {
	return clientset.MachinelearningV1().SeldonDeployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func CreateSeldonDeployment(ctx context.Context, deployment *v1.SeldonDeployment, namespace string) (*v1.SeldonDeployment, error) {
	return clientset.MachinelearningV1().SeldonDeployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
}

func UpdateSeldonDeployment(ctx context.Context, deployment *v1.SeldonDeployment, namespace string) (*v1.SeldonDeployment, error) {
	return clientset.MachinelearningV1().SeldonDeployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
}

//FIXME filter only to current deployment, what with the resyncing time?
func NewInformerFactory(namespace string) informer.SharedInformerFactory {
	return informer.NewSharedInformerFactoryWithOptions(clientset, time.Second*30, informer.WithNamespace(namespace))
}

//TODO config via file?
func getSeldonClientSet() (*seldonclientset.Clientset, error) {
	//var kubeconfig *string
	//if home := homedir.HomeDir(); home != "" {
	//	kubeconfig = flag.String("kubeconfig2", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	//} else {
	//	kubeconfig = flag.String("kubeconfig2", "", "absolute path to the kubeconfig file")
	//}
	//flag.Parse()

	config := ctrl.GetConfigOrDie()

	//config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	//if err != nil {
	//	return nil, err
	//}

	kubeClientset, err := seldonclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return kubeClientset, nil
}
