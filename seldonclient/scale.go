package seldonclient

import (
	"context"
	"fmt"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"k8s.io/client-go/util/retry"
)

func ScaleDeployment(ctx context.Context, name, namespace string, scale int) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, err := GetSeldonDeployment(ctx, name, namespace)
		if err != nil {
			return fmt.Errorf("failed to get latest version of deployment: %v", err)
		}
		updateDeploymentScale(deployment, scale)

		_, err = UpdateSeldonDeployment(ctx, deployment, namespace)
		return err
	})
	if retryErr != nil {
		return fmt.Errorf("scale operation failed: %v", retryErr)
	}
	return nil
}

// Arbitrary, just override any settings with general replica count
func updateDeploymentScale(deployment *v1.SeldonDeployment, scale int) {
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
