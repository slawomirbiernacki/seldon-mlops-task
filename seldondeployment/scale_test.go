package seldondeployment

import (
	"context"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	seldonfake "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestScale(t *testing.T) {

	ctx := context.Background()

	//given
	initialReplicas := int32(1)
	name := "test-dep"
	namespace := "test-namespace"
	deployment := v1.SeldonDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{},
		},
		Spec: v1.SeldonDeploymentSpec{
			Replicas: &initialReplicas,
			Predictors: []v1.PredictorSpec{
				{
					Replicas: &initialReplicas,
					ComponentSpecs: []*v1.SeldonPodSpec{
						{
							Replicas: &initialReplicas,
						},
						{
							Replicas: &initialReplicas,
						},
					},
				},
				{
					Replicas: &initialReplicas,
				},
			},
		},
	}
	targetReplicas := 2
	seldonClientset := seldonfake.NewSimpleClientset(&deployment)
	deploymentManager := NewManager(seldonClientset, namespace)

	//when
	err := deploymentManager.Scale(ctx, name, targetReplicas)
	if err != nil {
		t.Fatal(err)
	}

	//then
	updated, err := seldonClientset.MachinelearningV1().SeldonDeployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	helper := int32(targetReplicas)
	assertEqual(t, *updated.Spec.Replicas, helper)

	var nilPointer *int32 = nil
	for _, predictor := range updated.Spec.Predictors {
		assertEqual(t, predictor.Replicas, nilPointer)
		for _, comp := range predictor.ComponentSpecs {
			assertEqual(t, comp.Replicas, nilPointer)
		}
	}
}

// Normally would use some assertion library but here using this for simplicity
func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}
