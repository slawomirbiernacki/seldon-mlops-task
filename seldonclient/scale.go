package seldonclient

import (
	"context"
	"fmt"
	"k8s.io/client-go/util/retry"
)

func ScaleDeployment(ctx context.Context, name, namespace string, scale int) error {
	fmt.Printf("Scaling the deployment to %d replicas\n", scale)
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, err := GetSeldonDeployment(ctx, name, namespace)
		if err != nil {
			return fmt.Errorf("failed to get latest version of deployment: %v", err)
		}

		//TODO test, and is there a better way?
		replicas := int32(scale)
		//Override any nested replicas settings
		for i, _ := range deployment.Spec.Predictors {
			deployment.Spec.Predictors[i].Replicas = nil
			for j, _ := range deployment.Spec.Predictors[i].ComponentSpecs {
				deployment.Spec.Predictors[i].ComponentSpecs[j].Replicas = nil
			}
		}
		deployment.Spec.Replicas = &replicas

		_, err = UpdateSeldonDeployment(ctx, deployment, namespace)
		return err
	})
	if retryErr != nil {
		return fmt.Errorf("update failed: %v", retryErr)
	}
	fmt.Println("Scaling successful")
	return nil
}
