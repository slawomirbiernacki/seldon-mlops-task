package seldondeployment

import (
	"context"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	seldonclientset "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned"
	clientv1 "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned/typed/machinelearning.seldon.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// import all authentication handlers
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const DeletePropagationPolicy = metav1.DeletePropagationForeground

type Manager struct {
	seldonClientset seldonclientset.Interface
	deletePolicy    metav1.DeletionPropagation
	namespace       string
}

func NewManager(seldonClientset seldonclientset.Interface, namespace string) *Manager {
	manager := &Manager{
		seldonClientset: seldonClientset,
		deletePolicy:    DeletePropagationPolicy,
		namespace:       namespace,
	}
	return manager
}

func (manager *Manager) GetDeployment(ctx context.Context, name string) (*v1.SeldonDeployment, error) {
	return manager.getDeploymentInterface().Get(ctx, name, metav1.GetOptions{})
}

func (manager *Manager) DeleteDeployment(ctx context.Context, name string) (err error) {
	return manager.getDeploymentInterface().Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &manager.deletePolicy,
	})
}

func (manager *Manager) CreateDeployment(ctx context.Context, deployment *v1.SeldonDeployment) (*v1.SeldonDeployment, error) {
	return manager.getDeploymentInterface().Create(ctx, deployment, metav1.CreateOptions{})
}

func (manager *Manager) UpdateDeployment(ctx context.Context, deployment *v1.SeldonDeployment) (*v1.SeldonDeployment, error) {
	return manager.getDeploymentInterface().Update(ctx, deployment, metav1.UpdateOptions{})
}

func (manager *Manager) getDeploymentInterface() clientv1.SeldonDeploymentInterface {
	return manager.seldonClientset.MachinelearningV1().SeldonDeployments(manager.namespace)
}
