package fieldselectors

const AssignedNonTerminatedPods = "spec.nodeName!=,status.phase!=Failed,status.phase!=Succeeded"
const UnassignedNonTerminatedPods = "spec.nodeName=,status.phase!=Failed,status.phase!=Succeeded"
