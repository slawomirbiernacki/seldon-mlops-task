package seldondeployment

import (
	"context"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	seldonclientset "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned"
	clientv1 "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned/typed/machinelearning.seldon.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	// import all authentication handlers
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var seldonClientset seldonclientset.Interface
var clientset kubernetes.Interface
var deletePolicy = metav1.DeletePropagationForeground

func init() {
	config, err := ctrl.GetConfig()
	if err != nil {
		panic(err)
	}
	seldonClientset, err = seldonclientset.NewForConfig(config)

	if err != nil {
		panic(err)
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
}

func GetDeployment(ctx context.Context, name string, namespace string) (*v1.SeldonDeployment, error) {
	return deploymentInterface(namespace).Get(ctx, name, metav1.GetOptions{})
}

func DeleteDeployment(ctx context.Context, name string, namespace string) (err error) {
	return deploymentInterface(namespace).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

func CreateDeployment(ctx context.Context, deployment *v1.SeldonDeployment, namespace string) (*v1.SeldonDeployment, error) {
	return deploymentInterface(namespace).Create(ctx, deployment, metav1.CreateOptions{})
}

func UpdateDeployment(ctx context.Context, deployment *v1.SeldonDeployment, namespace string) (*v1.SeldonDeployment, error) {
	return deploymentInterface(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
}

func deploymentInterface(namespace string) clientv1.SeldonDeploymentInterface {
	return seldonClientset.MachinelearningV1().SeldonDeployments(namespace)
}
