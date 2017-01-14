package brain

import (
	"fmt"
	"strconv"

	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/util/uuid"
	utilrand "k8s.io/kubernetes/pkg/util/rand"

	"github.com/Prytu/risk-advisor/pkg/model"
)

func updateNewPodData(pod *v1.Pod, resourceVersion int64) {
	fillNewPodData(pod, resourceVersion, utilrand.String(model.MaxNameLength), unversioned.Now())
}

func fillNewPodData(pod *v1.Pod, resourceVersion int64, podName string, creationTimestamp unversioned.Time) {
	pod.UID = uuid.NewUUID()
	pod.CreationTimestamp = creationTimestamp
	if pod.Name == "" {
		pod.Name = podName
	}
	if pod.Namespace == "" {
		pod.Namespace = "default"
	}
	pod.SelfLink = fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", pod.Namespace, pod.Name)
	pod.ClusterName = ""
	pod.Status = v1.PodStatus{
		Phase: v1.PodPending,
	}
	pod.ResourceVersion = strconv.FormatInt(int64(resourceVersion), 10)
}

func bindPodToNode(pod *v1.Pod, nodeName string) {
	fillBoundPodData(pod, nodeName, unversioned.Now())
}

func fillBoundPodData(pod *v1.Pod, nodeName string, time unversioned.Time) {
	pod.Spec.NodeName = nodeName
	pod.Status.Conditions = []v1.PodCondition{
		{
			Type:               v1.PodScheduled,
			Status:             v1.ConditionTrue,
			LastTransitionTime: time,
		},
	}
}
