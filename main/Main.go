package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	seldonapi "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	seldonclientset "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned"
	informer "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/informers/externalversions"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var clientset *seldonclientset.Clientset

func init() {
	clientset, _ = GetSeldonClientSet()
}

func main() {
	namespace := "test-aaa"

	//FIXME understand resync time
	factory := informer.NewSharedInformerFactoryWithOptions(clientset, time.Second, informer.WithNamespace(namespace))
	events := make(chan struct{})

	//FIXME cache sync?

	informerr := factory.Machinelearning().V1().SeldonDeployments().Informer()
	defer close(events)

	go runSeldonCRDInformer(events, informerr, namespace)

	deployment := create(namespace)
	name := deployment.GetName()

	err := waitForDeployment(name, namespace, time.Second, 100*time.Second)
	if err != nil {
		panic(err)
	}

	get, err := GetSeldonDeployment(deployment.GetName(), namespace)
	if err != nil {
		panic(err)
	}
	fmt.Printf("status %s.\n", get.Status.State)

	err = DeleteSeldonDeployment(deployment.GetName(), namespace)
	if err != nil {
		panic(err)
	}
}

func runSeldonCRDInformer(stopCh <-chan struct{}, s cache.SharedIndexInformer, namespace string) {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			d := obj.(*seldonapi.SeldonDeployment)
			fmt.Printf("Added! %s.\n", d.GetName())

			// do what we want with the SeldonDeployment/event
		},
		DeleteFunc: func(obj interface{}) {
			d := obj.(*seldonapi.SeldonDeployment)
			fmt.Printf("Deleted! %s.\n", d.GetName())
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			d := newObj.(*seldonapi.SeldonDeployment)
			fmt.Printf("Updated! %s.\n", d.Status)
		},
	}
	s.AddEventHandler(handlers)
	s.Run(stopCh)
}

func create(namespace string) *seldonapi.SeldonDeployment {
	filename, _ := filepath.Abs("./main/test-resource.yaml")
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	d := &seldonapi.SeldonDeployment{}
	//deployment := &unstructured.Unstructured{}
	// Decode YAML to unstructured object.
	if _, _, err := decUnstructured.Decode(yamlFile, nil, d); err != nil {
		panic(err)
	}

	// try following https://erwinvaneyk.nl/kubernetes-unstructured-to-typed/
	//err = runtime.DefaultUnstructuredConverter.
	//	FromUnstructured(deployment.UnstructuredContent(), d)
	//if err != nil {
	//	fmt.Println("could not convert obj to SeldonDeployment")
	//	fmt.Print(err)
	//	return
	//}

	dep, err := CreateSeldonDeployment(d, namespace)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created deployment %q.\n", dep.GetName())

	return dep
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

func GetSeldonClientSet() (*seldonclientset.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig2", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig2", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	kubeClientset, err := seldonclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return kubeClientset, nil
}

func GetSeldonDeployment(name string, namespace string) (result *seldonapi.SeldonDeployment, err error) {
	return clientset.MachinelearningV1().SeldonDeployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func DeleteSeldonDeployment(name string, namespace string) (err error) {
	return clientset.MachinelearningV1().SeldonDeployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func CreateSeldonDeployment(deployment *seldonapi.SeldonDeployment, namespace string) (sdep *seldonapi.SeldonDeployment, err error) {
	return clientset.MachinelearningV1().SeldonDeployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
}

func waitForDeployment(name, namespace string, pollInterval, pollTimeout time.Duration) error {
	var reason string
	fmt.Print("Waiting\n")
	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		var err error
		deployment, err := GetSeldonDeployment(name, namespace)
		if err != nil {
			return false, err
		}

		// When the deployment status and its underlying resources reach the desired state, we're done
		if seldonapi.StatusStateAvailable == deployment.Status.State {
			//fmt.Print("\n")
			return true, nil
		}

		reason = fmt.Sprintf("deployment status: %#v", deployment.Status)

		//fmt.Print(".")
		return false, nil
	})

	if err == wait.ErrWaitTimeout {
		err = fmt.Errorf("%s", reason)
	}
	if err != nil {
		return fmt.Errorf("error waiting for deployment %q status to match expectation: %v", name, err)
	}
	return nil
}
