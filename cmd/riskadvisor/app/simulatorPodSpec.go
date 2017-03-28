package app

// TODO: move to ~config file

import "k8s.io/client-go/1.5/pkg/api/v1"

var simulatorPod = &v1.Pod{
	ObjectMeta: v1.ObjectMeta{
		Name: "simulator",
	},
	Spec: v1.PodSpec{
		Containers: []v1.Container{{
			Name:            "simulator",
			Image:           "pposkrobko/simulator",
			ImagePullPolicy: v1.PullNever, // only for development with Minikube
			Ports: []v1.ContainerPort{
				{ContainerPort: 9998},
				{ContainerPort: 9999},
			},
		},
			{
				Name:  "kubescheduler",
				Image: "gcr.io/google_containers/kube-scheduler:v1.4.6",
				Command: []string{"/bin/sh", "-c",
					"/usr/local/bin/kube-scheduler --master=127.0.0.1:9999 --leader-elect=false --kube-api-content-type application/json"},
			},
			{
				Name:            "kubectl",
				Image:           "gcr.io/google_containers/kubectl:v0.18.0-120-gaeb4ac55ad12b1-dirty",
				ImagePullPolicy: "Always",
				Args:            []string{"proxy", "-p", "8080"},
			},
		},
	},
}
