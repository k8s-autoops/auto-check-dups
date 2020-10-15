# auto-check-dups

## Usage

Create namespace `autoops` and apply yaml resources as described below.

```yaml
# create serviceaccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: auto-check-dups
  namespace: autoops
---
# create clusterrole
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: auto-check-dups
rules:
  - apiGroups: [""]
    resources: ["services"]
    verbs: ["list"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets"]
    verbs: ["list"]
---
# create clusterrolebinding
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: auto-check-dups
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: auto-check-dups
subjects:
  - kind: ServiceAccount
    name: auto-check-dups
    namespace: autoops
---
# create cronjob
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: auto-check-dups
  namespace: autoops
spec:
  schedule: "*/5 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccount: auto-check-dups
          containers:
            - name: auto-check-dups
              image: autoops/auto-check-dups
          restartPolicy: OnFailure
```

## Credits

Guo Y.K., MIT License
