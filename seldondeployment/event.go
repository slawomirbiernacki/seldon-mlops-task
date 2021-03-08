package seldondeployment

import (
	seldonv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var scheme = runtime.NewScheme()

func init() {
	err := seldonv1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
}

type EventProcessor struct {
	seen       map[types.UID]bool
	deployment *seldonv1.SeldonDeployment
	namespace  string
}

// Fetches a batch of events and processes them. Function is incremental, skipping any previously seen events.
// Event version is ignored - event identity is based solely on UUID, therefore any given event is processed only once.
func (ep *EventProcessor) ProcessNew(eventProcessor func(event corev1.Event) error) error {
	events, err := clientset.CoreV1().Events(ep.namespace).Search(scheme, ep.deployment)
	if err != nil {
		return err
	}

	for _, event := range events.Items {
		if !ep.seen[event.UID] {
			err := eventProcessor(event)
			if err != nil {
				return err
			}
			ep.seen[event.UID] = true
		}
	}
	return nil
}

// Stateful event processor. It has a cache to remember processed events. Quite naive and probably not good idea for production code, pragmatically assumed here it's ok.
func NewEventProcessor(deployment *seldonv1.SeldonDeployment, namespace string) *EventProcessor {
	watcher := &EventProcessor{
		seen:       make(map[types.UID]bool),
		deployment: deployment,
		namespace:  namespace,
	}
	return watcher
}
