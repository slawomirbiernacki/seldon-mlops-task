package seldonclient

import (
	"context"
	"fmt"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

const PollInterval = time.Second

//FIXME flag for timeout
func WaitForDeploymentStatus(ctx context.Context, deploymentName, namespace string, status v1.StatusState, pollTimeout time.Duration) error {

	fmt.Printf("Waiting for status: %s of deployment %s \n", status, deploymentName)
	err := wait.PollImmediate(PollInterval, pollTimeout, func() (bool, error) {
		var err error
		deployment, err := GetSeldonDeployment(ctx, deploymentName, namespace)
		if err != nil {
			return false, err
		}

		// When the deployment status and its underlying resources reach the desired state, we're done
		if status == deployment.Status.State {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return fmt.Errorf("error waiting for deployment %q status to match expectation: %v", deploymentName, err)
	}

	fmt.Printf("Deployment %s reached status: %v\n", deploymentName, status)
	return nil
}

//TODO name
func WaitForScale(ctx context.Context, deploymentName, namespace string, scale int, pollTimeout time.Duration) error {

	fmt.Printf("Waiting for deployment %s to reach target replica count\n", deploymentName)
	err := wait.PollImmediate(PollInterval, pollTimeout, func() (bool, error) {
		var err error
		deployment, err := GetSeldonDeployment(ctx, deploymentName, namespace)
		if err != nil {
			return false, err
		}

		replicas := int32(scale)
		finished := true
		for _, deploymentStatus := range deployment.Status.DeploymentStatus {
			// Quite arbitrary, would need to have more kubernetes knowledge to be confident this is the right assertion
			if deploymentStatus.AvailableReplicas != replicas && deploymentStatus.Replicas != replicas {
				finished = false
			}
		}
		return finished, nil
	})

	if err != nil {
		return fmt.Errorf("error waiting for deployment %q replicas to match expectation: %v", deploymentName, err)
	}

	fmt.Printf("Deployment %s reached target replica count\n", deploymentName)
	return nil
}

func WaitUntilDeploymentDeleted(ctx context.Context, deploymentName, namespace string, pollTimeout time.Duration) error {

	err := wait.PollImmediate(PollInterval, pollTimeout, func() (bool, error) {
		var err error
		_, err = GetSeldonDeployment(ctx, deploymentName, namespace)

		if err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}

			return false, err
		}

		return false, nil
	})

	if err != nil {
		return err
	}

	return nil
}
