package brain

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/client-go/1.5/pkg/api/v1"
	metav1 "k8s.io/client-go/1.5/pkg/apis/meta/v1"
	utilrand "k8s.io/kubernetes/pkg/util/rand"
)

func TestUpdateNewEmptyPodData(t *testing.T) {
	pod := &v1.Pod{}
	resourceVersion := int64(1)
	time := metav1.Now()
	name := utilrand.String(20)

	fillNewPodData(pod, resourceVersion, name, time)

	assert.Empty(t, pod.ClusterName)
	assert.NotEmpty(t, string(pod.UID))
	assert.Equal(t, name, pod.Name)
	assert.Equal(t, "default", pod.Namespace)
	assert.Equal(t, time, pod.CreationTimestamp)
	assert.Equal(t, strconv.Itoa(1), pod.ResourceVersion)
	assert.Equal(t, v1.PodStatus{Phase: v1.PodPending}, pod.Status)
	assert.Equal(t, fmt.Sprintf("/api/v1/namespaces/default/pods/%s", pod.Name), pod.SelfLink)
}

func TestUpdateNewPodData(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      "pod",
			Namespace: "namespace",
		},
	}
	resourceVersion := int64(1)
	time := metav1.Now()

	fillNewPodData(pod, resourceVersion, "", time)

	assert.NotEmpty(t, pod.UID)
	assert.Empty(t, pod.ClusterName)
	assert.Equal(t, "pod", pod.Name)
	assert.Equal(t, "namespace", pod.Namespace)
	assert.Equal(t, time, pod.CreationTimestamp)
	assert.Equal(t, strconv.Itoa(1), pod.ResourceVersion)
	assert.Equal(t, v1.PodStatus{Phase: v1.PodPending}, pod.Status)
	assert.Equal(t, fmt.Sprintf("/api/v1/namespaces/namespace/pods/%s", pod.Name), pod.SelfLink)
}

func TestBindPodToNode(t *testing.T) {
	pod := &v1.Pod{}
	nodeName := "nodename"
	time := metav1.Now()

	fillBoundPodData(pod, nodeName, time)

	assert.Equal(t, "nodename", pod.Spec.NodeName)
	expectedPodCondition := []v1.PodCondition{{Type: v1.PodScheduled, Status: v1.ConditionTrue, LastTransitionTime: time}}
	assert.Equal(t, expectedPodCondition, pod.Status.Conditions)
}
