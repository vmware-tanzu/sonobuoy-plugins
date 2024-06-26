def checkPods(pods, deployment):
    isClusterFullForAnyPods(pods)
    checkIfQuotaLimiting(deployment)
    checkPVCs(pods)
    checkAssignedToNodes(pods)

def checkAssignedToNodes(pods):
  sonobuoy.startTest("Pods should be assigned to nodes")
  for pod in pods:
    err = assignedToNode(pod)
    if err != "":
      sonobuoy.failTest(err)
      return
  sonobuoy.passTest()
  sonobuoy.startTest("Pending pod status should be resolved")
  sonobuoy.failTest("Unknown reason why pods are in the unready status. There may be a problem with the kubelet")

def assignedToNode(pod):
  if not hasattr(pod.spec, "nodeName"):
    return "The pod " + pod.metadata.name + " is not assigned to a node. There may be a problem with the scheduler."
  return ""

# Returns an error to report if pvc is pending
def checkPVCs(pods):
 sonobuoy.startTest("PVC should not be pending")
 for pod in pods:
   err = checkPVC(pod)
   if err != "":
     sonobuoy.failTest(err)
     return
 sonobuoy.passTest()

def checkPVC(pod):
  ns = pod.metadata.namespace
  for v in pod.spec.volumes:
    if hasattr(v,"persistentVolumeClaim"):
      name = v.persistentVolumeClaim.claimName
      if not kube.exists(persistentVolumeClaim=ns+"/"+name):
        return "Claim " + name + " does not exist create it or fix the PVC reference in the deployment"
      pvcStr = kube.get(persistentVolumeClaim=ns+"/"+name)
      pvc = sToObj(pvcStr)
      if pvc.status.phase == "Pending":
        return "Claim " + name + " is in the pending phase and may have issues with its definition"
  return ""

def isClusterFullForAnyPods(pods):
    sonobuoy.startTest("Cluster has sufficient space for pods")
    for pod in pods:
      if pod.status.phase == "Pending":
        if isClusterFull(pod):
          sonobuoy.failTest("Provision a larger cluster or reduce the number of replicas in the deployment")
          return
    sonobuoy.passTest()

# Checks a pod to see if it seems the cluster is too full. Based on error messages so it is somewhat rigid.
def isClusterFull(pod):
  for cond in pod.status.conditions:
    if cond.type == "PodScheduled" and cond.status == "False" and cond.reason == "Unschedulable" and cond.message.count("Too many pods") > 0:
      print("Pod " + pod.metadata.name + " is unschedulable with the following message set: " + cond.message)
      return True
  return False

def checkIfQuotaLimiting(deployment):
  sonobuoy.startTest("Resource Quotas respected")
  for cond in deployment.status.conditions:
    if cond.type == "ReplicaFailure" and cond.status == "True" and cond.reason == "FailedCreate":
      sonobuoy.failTest("Deployment exceeds quota. Currently has a ReplicaFailure with the message: " + cond.message + ". Increase quotas or decrease deployment requirements.")
      return
  sonobuoy.passTest()

def getPendingPods(ns):
    podsString = kube.get(pod=ns+"/")
    podsDecode = yaml.decode(podsString)
    pods = struct.encode(podsDecode)
    return pods.items

def checkService(service, deployment):
  d = sToObj(deployment)
  s = sToObj(service)
  assert.equals(s.spec.selector, d.spec.template.metadata.labels, "Service labels and deployment labels should match but got diff: $3. Update one of them to fix the issue.")
  sonobuoy.passTest()
  sonobuoy.startTest("Service references proper pod labels")
  assert.equals(s.spec.selector, d.spec.template.metadata.labels, "Service labels and deployment labels should match but got diff: $3. Update one of them to fix the issue.")
  sonobuoy.passTest()
  sonobuoy.startTest("Service and deployment agree on port to communicate on")
  assert.equals(s.spec.ports[0].targetPort, d.spec.template.spec.containers[0].ports[0].containerPort, "Ingress refers to service port $1 but the service uses $2. Update one of them so they match.")
  sonobuoy.passTest()
  if kube.exists(endpoints=env.service):
      checkServiceReachability()

def checkIngress(ingress, service):
  i = sToObj(ingress)
  s = sToObj(service)
  sonobuoy.startTest("Ingress refers to proper service name")
  assert.equals(i.spec.rules[0].http.paths[0].backend.service.name, s.metadata.name, "Ingress refers to service $1 but the name should be $2. Update one of them so that they match.")
  sonobuoy.passTest()
  sonobuoy.startTest("Ingress refers to proper service port")
  assert.equals(i.spec.rules[0].http.paths[0].backend.service.port.number, s.spec.ports[0].port, "Ingress refers to service port $1 but the service uses $2. Update one of them so they match.")
  sonobuoy.passTest()

def checkDeployment(deployment):
  d = sToObj(deployment)
  sonobuoy.startTest("Deployment pods should have the same labels and matchLabel")
  assert.equals(d.spec.selector.matchLabels, d.spec.template.metadata.labels, "Deployment matchLabel should always match the labels in the template spec but got diff: $3. Update one of them to fix the issue.")
  sonobuoy.passTest()

def checkServiceReachability():
  print("TODO checks for port-forwarding and reachability")

def sToObj(s):
  return struct.encode(yaml.decode(s))

def getPods(deployment):
  ns = env.deployment.split("/")[0]
  d = sToObj(deployment)
  appLabel = d.spec.selector.matchLabels
  lbls=list()
  for k in appLabel:
    lbls.append(k + "=" + appLabel[k])
  lblSelector = ",".join(lbls)
  pl = kube.get(pods=ns+"/?labelSelector="+lblSelector)
  return sToObj(pl).items

sonobuoy.startSuite()

i = kube.get(ingress=env.ingress)
s = kube.get(service=env.service)
d = kube.get(deployment=env.deployment)

p = getPods(d)
checkPods(p, sToObj(d))
checkDeployment(d)
checkService(s,d)
checkIngress(i,s)
sonobuoy.done()
