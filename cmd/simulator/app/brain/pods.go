package brain

import (
	"fmt"
	"strconv"

	"k8s.io/client-go/pkg/api/v1"
	metav1 "k8s.io/client-go/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/util/uuid"
	utilrand "k8s.io/kubernetes/pkg/util/rand"

	"github.com/Prytu/risk-advisor/pkg/model"
)

func updateNewPodData(pod *v1.Pod, resourceVersion int64) {
	fillNewPodData(pod, resourceVersion, utilrand.String(model.MaxNameLength), metav1.Now())
}

func fillNewPodData(pod *v1.Pod, resourceVersion int64, podName string, creationTimestamp metav1.Time) {
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
	fillBoundPodData(pod, nodeName, metav1.Now())
}

func fillBoundPodData(pod *v1.Pod, nodeName string, time metav1.Time) {
	pod.Spec.NodeName = nodeName
	pod.Status.Conditions = []v1.PodCondition{
		{
			Type:               v1.PodScheduled,
			Status:             v1.ConditionTrue,
			LastTransitionTime: time,
		},
	}
}
