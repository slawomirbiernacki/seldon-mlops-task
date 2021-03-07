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

// This is quite arbitrary, just override any settings with general replica count
func updateDeploymentScale(deployment *v1.SeldonDeployment, scale int) {
	//TODO test, and is there a better way?
	replicas := int32(scale)
	//Override any nested replica settings
	for i := range deployment.Spec.Predictors {
		deployment.Spec.Predictors[i].Replicas = nil
		for j := range deployment.Spec.Predictors[i].ComponentSpecs {
			deployment.Spec.Predictors[i].ComponentSpecs[j].Replicas = nil
		}
	}
	deployment.Spec.Replicas = &replicas
}
