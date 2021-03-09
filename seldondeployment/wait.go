package seldondeployment

import (
	"context"
	"fmt"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

const PollInterval = time.Second

func (manager *Manager) WaitUntilDeploymentStatus(ctx context.Context, deploymentName string, status v1.StatusState, pollTimeout time.Duration) error {

	condition := func(deployment *v1.SeldonDeployment, err error) (bool, error) {
		if err != nil {
			return false, err
		}

		if status == deployment.Status.State {
			return true, nil
		}

		return false, nil
	}
	return waitUntilCondition(ctx, manager, deploymentName, condition, pollTimeout)
}

func (manager *Manager) WaitUntilDeploymentScaled(ctx context.Context, deploymentName string, scale int, pollTimeout time.Duration) error {

	condition := func(deployment *v1.SeldonDeployment, err error) (bool, error) {
		if err != nil {
			return false, err
		}

		replicas := int32(scale)
		finished := true
		for _, deploymentStatus := range deployment.Status.DeploymentStatus {
			// Arbitrary, not entirely sure this is the right assertion
			if deploymentStatus.AvailableReplicas != replicas && deploymentStatus.Replicas != replicas {
				finished = false
			}
		}
		return finished, nil
	}
	return waitUntilCondition(ctx, manager, deploymentName, condition, pollTimeout)
}

func (manager *Manager) WaitUntilDeploymentDeleted(ctx context.Context, deploymentName string, pollTimeout time.Duration) error {

	condition := func(deployment *v1.SeldonDeployment, err error) (bool, error) {
		if err != nil && errors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}

	return waitUntilCondition(ctx, manager, deploymentName, condition, pollTimeout)
}

func waitUntilCondition(ctx context.Context, manager *Manager, deploymentName string, condition func(deployment *v1.SeldonDeployment, err error) (bool, error), pollTimeout time.Duration) error {

	err := wait.PollImmediate(PollInterval, pollTimeout, func() (bool, error) {
		deployment, err := manager.GetDeployment(ctx, deploymentName)
		return condition(deployment, err)
	})

	if err != nil {
		return fmt.Errorf("error while waiting for %s: %w", deploymentName, err)
	}
	return nil
}
