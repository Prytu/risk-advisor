package app

import (
	"k8s.io/kubernetes/pkg/api/v1"
	metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"
)

func PodListFromPod(pod *v1.Pod) *v1.PodList {
	typeMeta := metav1.TypeMeta{
		Kind:       "PodList",
		APIVersion: "v1",
	}

	listMeta := metav1.ListMeta{
		SelfLink:        "/api/v1/pods",
		ResourceVersion: "1",
	}

	return &v1.PodList{
		TypeMeta: typeMeta,
		ListMeta: listMeta,
		Items:    []v1.Pod{*pod},
	}
}

func EmptyPodList() *v1.PodList {
	typeMeta := metav1.TypeMeta{
		Kind:       "PodList",
		APIVersion: "v1",
	}

	listMeta := metav1.ListMeta{
		SelfLink:        "/api/v1/pods",
		ResourceVersion: "1",
	}

	return &v1.PodList{
		TypeMeta: typeMeta,
		ListMeta: listMeta,
		Items:    make([]v1.Pod, 0),
	}
}
