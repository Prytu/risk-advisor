package app

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

func PodListFromPod(pod *api.Pod) *api.PodList {
	typeMeta := unversioned.TypeMeta{
		Kind:       "PodList",
		APIVersion: "v1",
	}

	listMeta := unversioned.ListMeta{
		SelfLink:        "/api/v1/pods",
		ResourceVersion: "1",
	}

	return &api.PodList{
		TypeMeta: typeMeta,
		ListMeta: listMeta,
		Items:    []api.Pod{*pod},
	}
}

func EmptyPodList() *api.PodList {
	typeMeta := unversioned.TypeMeta{
		Kind:       "PodList",
		APIVersion: "v1",
	}

	listMeta := unversioned.ListMeta{
		SelfLink:        "/api/v1/pods",
		ResourceVersion: "1",
	}

	return &api.PodList{
		TypeMeta: typeMeta,
		ListMeta: listMeta,
		Items:    make([]api.Pod, 0),
	}
}
