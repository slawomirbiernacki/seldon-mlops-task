package seldonclient

import (
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"path/filepath"
)

func ParseDeploymentFromFile(path string) (*v1.SeldonDeployment, error) {
	filename, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	var decodingSerializer = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	deployment := &v1.SeldonDeployment{}

	if _, _, err := decodingSerializer.Decode(yamlFile, nil, deployment); err != nil {
		return nil, err
	}
	return deployment, nil
}
