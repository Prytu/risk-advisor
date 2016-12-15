package urls

const AssignedNonTerminatedPods = "/api/v1/pods?fieldSelector=spec.nodeName!=,status.phase!=Failed,status.phase!=Succeeded&resourceVersion=0"
const UnassignedNonTerminatedPods = "/api/v1/pods?fieldSelector=spec.nodeName=,status.phase!=Failed,status.phase!=Succeeded&resourceVersion=0"
const Nodes = "/api/v1/nodes?resourceVersion=0"
const Pvcs = "/api/v1/persistentvolumeclaims?resourceVersion=0"
const Pvs = "/api/v1/persistentvolumes?resourceVersion=0"
const Replicasets = "/apis/extensions/v1beta1/replicasets?resourceVersion=0"
const Services = "/api/v1/services?resourceVersion=0"
const ReplicationControllers = "/api/v1/replicationcontrollers?resourceVersion=0"
