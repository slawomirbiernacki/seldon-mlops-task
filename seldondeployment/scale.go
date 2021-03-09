package seldondeployment

import (
	"context"
	"fmt"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"k8s.io/client-go/util/retry"
)

func (manager *Manager) Scale(ctx context.Context, name string, scale int) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, err := manager.GetDeployment(ctx, name)
		if err != nil {
			return fmt.Errorf("failed to get latest version of deployment: %v", err)
		}
		updateScale(deployment, scale)

		_, err = manager.UpdateDeployment(ctx, deployment)
		return err
	})
	if retryErr != nil {
		return fmt.Errorf("scale operation failed: %v", retryErr)
	}
	return nil
}

// Arbitrary logic, just override any settings with general replica count. Ignoring svcOrchSpec.replicas .
func updateScale(deployment *v1.SeldonDeployment, scale int) {
	replicas := int32(scale)
	deployment.Spec.Replicas = &replicas

	//Unset any nested replica settings to make update effective
	for predIdx := range deployment.Spec.Predictors {
		deployment.Spec.Predictors[predIdx].Replicas = nil
		for compIdx := range deployment.Spec.Predictors[predIdx].ComponentSpecs {
			deployment.Spec.Predictors[predIdx].ComponentSpecs[compIdx].Replicas = nil
		}
	}
}
