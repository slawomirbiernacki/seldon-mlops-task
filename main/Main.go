package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	//"gopkg.in/yaml.v3"

	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	//deploymentutil "k8s.io/kubernetes/pkg/controller/deployment/util"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	//// create the clientset
	//clientset, err := kubernetes.NewForConfig(config)
	//if err != nil {
	//	panic(err.Error())
	//}
	//
	//pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	//if err != nil {
	//	panic(err.Error())
	//}
	//
	//for _, svc:=range pods.Items{
	//	fmt.Fprintf(os.Stdout, "service name: %v\n", svc.Name)
	//}
	//
	//fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	//main2.Test(config)
	create(config)
}

func create(config *rest.Config) {
	filename, _ := filepath.Abs("./main/test-resource.yaml")
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	namespace := "test-aaa"

	var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	deployment := &unstructured.Unstructured{}
	// Decode YAML to unstructured object.
	if _, _, err := decUnstructured.Decode(yamlFile, nil, deployment); err != nil {
		panic(err)
	}

	//var dep seldon_protos.SeldonDeployment
	//
	//err = yaml.Unmarshal(yamlFile, &dep)
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Printf("Value: %#v\n", deployment)

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deploymentRes := schema.GroupVersionResource{Group: "machinelearning.seldon.io", Version: "v1", Resource: "seldondeployments"}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := client.Resource(deploymentRes).Namespace("test-aaa").Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetName())

	status, found, err := unstructured.NestedStringMap(result.Object, "status")
	fmt.Printf("status %q.\n", status)
	fmt.Printf("found %q.\n", found)
	//fmt.Printf("status %q.\n", result)

	//clientset, err := kubernetes.NewForConfig(config)
	//if err != nil {
	//	panic(err.Error())
	//}
	//
	//deployments, err := clientset.AppsV1().Deployments("test-aaa").Get(context.TODO(), result.GetName(),  metav1.GetOptions{})
	//if err != nil {
	//	panic(err.Error())
	//}

	// Delete Deployment

	watchEvents(config)
	prompt()

	// List Deployments
	fmt.Printf("Listing deployments in namespace ")
	get, err := client.Resource(deploymentRes).Namespace(namespace).Get(context.TODO(), result.GetName(), metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	replicas, found, err := unstructured.NestedInt64(get.Object, "spec", "replicas")
	if err != nil || !found {
		fmt.Printf("Replicas found for deployment %s: error=%s", replicas, err)
	}

	prompt()

	fmt.Println("Deleting deployment...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := client.Resource(deploymentRes).Namespace("test-aaa").Delete(context.TODO(), result.GetName(), metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted deployment.")
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

func watchEvents(config *rest.Config) {
	//ctx, _ := context.WithCancel(context.Background()) //cancel!
	////defer cancel()
	//// create the clientset
	//client, err := kubernetes.NewForConfig(config)
	//if err != nil {
	//	panic(err.Error())
	//}
	//
	//deploymentRes := schema.GroupVersionResource{Group: "machinelearning.seldon.io", Version: "v1", Resource: "seldondeployments"}

	// We will create an informer that writes added pods to a channel.
	//pods := make(chan *v1.Pod, 1)
	//informers := informers.NewSharedInformerFactory(client, 0)
	//podInformer, err := informers.Core().V1().ForResource(deploymentRes)
	//if err != nil {
	//	panic(err.Error())
	//}
	//podInformer.Informer().AddEventHandler(&cache.ResourceEventHandlerFuncs{
	//	AddFunc: func(obj interface{}) {
	//		pod := obj.(*v1.Service)
	//		fmt.Printf("pod added: %s/%s\n", pod.Namespace, pod.Name)
	//		//pods <- pod
	//	},
	//})
	//
	//// Make sure informers are running.
	//informers.Start(ctx.Done())

	//prompt()

	//for elem := range pods {
	//	fmt.Println(elem)
	//}
}

//func waitForDeploymentCompleteMaybeCheckRolling(c kubernetes.Clientset, d *apps.Deployment , pollInterval, pollTimeout time.Duration) error {
//	var (
//		deployment *apps.Deployment
//		reason     string
//	)
//
//	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
//		var err error
//		deployment, err = c.AppsV1().Deployments("test-aaa").Get(context.TODO(), d.Name, metav1.GetOptions{})
//		if err != nil {
//			return false, err
//		}
//
//		// If during a rolling update, make sure rolling update strategy isn't broken at any times.
//
//		// When the deployment status and its underlying resources reach the desired state, we're done
//		if deploymentutil.DeploymentComplete(d, &deployment.Status) {
//			return true, nil
//		}
//
//		reason = fmt.Sprintf("deployment status: %#v", deployment.Status)
//
//		return false, nil
//	})
//
//	if err == wait.ErrWaitTimeout {
//		err = fmt.Errorf("%s", reason)
//	}
//	if err != nil {
//		return fmt.Errorf("error waiting for deployment %q status to match expectation: %v", d.Name, err)
//	}
//	return nil
//}
