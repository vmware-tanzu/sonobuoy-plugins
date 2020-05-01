# who-can

This plugin utilizes the [`kubectl-who-can` project from AquaSecurity](https://github.com/aquasecurity/kubectl-who-can) to produce a variety of reports to show which subjects have RBAC permissions to perform actions (verbs) against resources in the cluster.

This allows you to audit what actions Subjects in your cluster can perform and also view which Subjects can perform particular actions.
By providing data in both of these views, you can inspect the RBAC permissions for a particular Subject or for a particular resource.

This plugin makes use of a small runner which finds all the API resources available in the cluster.
It then iterates over all of these resources and subresources and checks which subjects can perform each of the supported verbs for the resource.

By default, it will perform the check against the default namespace.
This means that if the query is to check who can `create pods`, it will only check who can create pods in the default namespace.
Additional namespaces to query against can be specified by modifying the `WHO_CAN_CONFIG` entry in the [plugin definition](./who-can.yaml) to add more namespaces to the list.
The plugin definition currently includes the `kube-system` namespace and "all namespaces" (`*`) in this list.

## Prerequisites
This plugin uses features that are only available in [Sonobuoy v0.18.1](https://github.com/vmware-tanzu/sonobuoy/releases/tag/v0.18.1).
It will still run with earlier versions of Sonobuoy however its overall `status` will be incorrectly reported.

## Usage

To run this plugin, run the following command:

```
sonobuoy run --plugin https://raw.githubusercontent.com/vmware-tanzu/sonobuoy-plugins/master/who-can/who-can.yaml
```

The plugin status can be checked using the command:

```
sonobuoy status
```

Once the plugin is complete, retrieve the results using the command:

```
sonobuoy retrieve
```

This command will return the name of the results tarball.

By default, the plugin produces three report files, each displaying the same information but in different formats and views.
These can be found in the tarball at the following paths:
 *  `plugins/who-can/results/global/subjects-report.json`
 *  `plugins/who-can/results/global/resources-report.json`
 *  `plugins/who-can/results/global/sonobuoy_results.yaml`

Each of these different files is explained below.

## Report formats

The plugin is capable of producing three different report formats or views:
 * Subjects report (`--subjects-report`)
 * Resources report (`--resources-report`)
 * Sonobuoy report (`--sonobuoy-report`)
 
### Subjects report

This report is a view of the RBAC data in the cluster grouped by subjects, detailing what actions each subject can perform in each of the queried namespaces.
For each action that a subject can perform, the RoleBindings and ClusterRoleBindings that allow that subject to perform an action are listed.

This report is a JSON file with the following format:

```
[
  {
    "kind": "ServiceAccount",
    "name": "expand-controller",
    "namespace": "kube-system",
    "permissions": [
      {
        "namespace": "*",
        "actions": [
          {
            "resource": "endpoints",
            "verb": "get",
            "cluster-role-bindings": [
              "system:controller:expand-controller"
            ]
          },
          {
            "resource": "events",
            "verb": "create",
            "cluster-role-bindings": [
              "system:controller:expand-controller"
            ]
          },
          ...
        ]
      },
      {
        "namespace": "kube-system",
        "actions": [
          {
            "resource": "endpoints",
            "verb": "get",
            "cluster-role-bindings": [
              "system:controller:expand-controller"
            ]
          },
          {
            "resource": "events",
            "verb": "create",
            "cluster-role-bindings": [
              "system:controller:expand-controller"
            ]
          },
        ]
      },
    ],
  },
  {
    "kind": "User",
    "apiGroup": "rbac.authorization.k8s.io",
    "name": "pod-lister",
    "permissions": [
      {
        "namespace": "secret",
        "actions": [
          {
            "resource": "pods",
            "verb": "list",
            "role-bindings": [
              "list-secret-pods"
            ]
          }
        ]
      }
    ]
  },
  ...
}
```

As we can see in the above example, the ServiceAccount `expand-controller`, which is in the namespace `kube-system`, has permissions to `get endpoints` in all namespaces (`*`) due to the `system:controller:expand-contoller` cluster role binding.

We can also see that the User `pod-lister`, only has permissions to `list pods` in the `secret` namespace due to the `list-secret-pods` RoleBinding.

### Resources report

This report is a view of the RBAC data in the cluster, detailing which subjects can performs actions against a resource in a particular namespace.
Along with each subject are the RoleBindings and ClusterRoleBindings that allow them to perform that action.

This report is a JSON file with the following format:

```
  {
    "resource": "pods",
    "verb": "list",
    "namespace": "secret",
    "subjects": [
      {
        "kind": "ServiceAccount",
        "name": "pvc-protection-controller",
        "namespace": "kube-system",
        "cluster-role-bindings": [
          "system:controller:pvc-protection-controller"
        ]
      },
      {
        "kind": "ServiceAccount",
        "name": "replication-controller",
        "namespace": "kube-system",
        "cluster-role-bindings": [
          "system:controller:replication-controller"
        ]
      },
      {
        "kind": "User",
        "apiGroup": "rbac.authorization.k8s.io",
        "name": "pod-lister",
        "role-bindings": [
          "list-secret-pods"
        ]
      },
      {
        "kind": "ServiceAccount",
        "name": "node-controller",
        "namespace": "kube-system",
        "cluster-role-bindings": [
          "system:controller:node-controller"
        ]
      },

      ...
    ]
  }
```

In the above example, we can see four subjects that have the ability to `list pods` in the `secret` namespace.
The ServiceAccount `pvc-protection-controller`, which is in the namespace `kube-system`, can perform this action due to the `system:controller:pvc-protection-controller` ClusterRoleBinding.
Like in the Subjects report example above, we can see again that the User `pod-lister` has permission to perform this action due to the `list-secret-pods` RoleBinding.

### Sonobuoy report

This report is a variant of the Resources report using the [Sonobuoy Results format](https://sonobuoy.io/simplified-results-reporting-with-sonobuoy/).
This enables the plugin results to be processed and presented easily through Sonobuoy.

This report is a YAML file with the following format:

```
name: who-can
status: complete
items:
- name: create bindings -n kube-system
  details:
    namespace: kube-system
    resource: bindings
    verb: create
  items:
  - name: system:masters
    details:
      kind: Group
    items:
    - name: clusterrolebindings
      items:
      - name: cluster-admin
  - name: sonobuoy-serviceaccount
    details:
      kind: ServiceAccount
      namespace: sonobuoy
    items:
    - name: clusterrolebindings
      items:
      - name: sonobuoy-serviceaccount-sonobuoy
  ...
```

The Sonobuoy results format is used to describe the results for a plugin.
In this report, we can see that the `who-can` plugin has the status `complete`.
The `items` entry is an array where each entry represents a check that was performed.
The first item was the check to find which Subjects could `create bindings` in the `kube-system` namespace.
In the `details` map, we can see the individual components that made up this check.
The result `items` for this check are the Subjects that are permitted to perform this action along with the `rolebindings` or `clusterrolebindings` that enable them to perform the action.
