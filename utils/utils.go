package utils

import (
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"path/filepath"
)

var decodingSerializer = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

func ParseDeploymentFromFile(path string, into runtime.Object) error {
	filename, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if _, _, err := decodingSerializer.Decode(yamlFile, nil, into); err != nil {
		return err
	}
	return nil
}
