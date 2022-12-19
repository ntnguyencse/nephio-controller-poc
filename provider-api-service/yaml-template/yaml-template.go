package yamlFileTemplate

const (
	JobsTemplate = `apiVersion: batch/v1
kind: Job
metadata:
  name: placeholder-name
  namespace: nephio-system
spec:
  ttlSecondsAfterFinished: 100
  template:
    spec:
      nodeSelector:
        kubernetes.io/hostname=hyfast-mgmt01
      containers:
      - name: job
        image: dcnstarlab/job-cluster:latest
        env:
        // - name: OBJECT_NAME, STATUS, LAyer
        - name: CLUSTER_NAME
          value: placeholder-cluster-name
        - name: CLUSTER_NAMESPACE
          value: placeholder-cluster-namespace
        volumeMounts:
        - name: provider-configmap
          mountPath: /etc/openstack/clouds.yaml
          subPath: clouds.yaml
        - name: emco-configmap
          mountPath: /workspace/emco-cfg.yaml
          subPath: emco-cfg.yaml
        - name: kubeconfig
          mountPath: /kubeconfig
      volumes:
      - name: provider-configmap
        secret:
          secretName: openstack-admin
          items:
          - key: clouds.yaml
            path: clouds.yaml
      - name: emco-configmap
        configMap:
          name: emco-env
          items:
          - key: emco-cfg.yaml
            path: emco-cfg.yaml
      - name: kubeconfig
        configMap:
          name: clusterapi-management-kubeconfig
          defaultMode: 420
      restartPolicy: Never
`
)
