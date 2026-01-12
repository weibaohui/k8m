import React, { useState } from 'react';
import { Button, Dropdown } from 'antd';
import { DownOutlined } from '@ant-design/icons';

const BUILTIN_TEMPLATES = {
  workload: {
    label: 'Workload',
    children: {
      deployment: {
        label: 'Deployment',
        content: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: my-app
        image: nginx:latest
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"`
      },
      statefulset: {
        label: 'StatefulSet',
        content: `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: my-statefulset
  namespace: default
spec:
  serviceName: my-statefulset-service
  replicas: 3
  selector:
    matchLabels:
      app: my-statefulset
  template:
    metadata:
      labels:
        app: my-statefulset
    spec:
      containers:
      - name: my-container
        image: nginx:latest
        ports:
        - containerPort: 80
        volumeMounts:
        - name: data
          mountPath: /data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 1Gi`
      },
      daemonset: {
        label: 'DaemonSet',
        content: `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: my-daemonset
  namespace: default
spec:
  selector:
    matchLabels:
      app: my-daemonset
  template:
    metadata:
      labels:
        app: my-daemonset
    spec:
      containers:
      - name: my-container
        image: nginx:latest
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"`
      },
      pod: {
        label: 'Pod',
        content: `apiVersion: v1
kind: Pod
metadata:
  name: my-pod
  namespace: default
spec:
  containers:
  - name: my-container
    image: nginx:latest
    ports:
    - containerPort: 80
    resources:
      requests:
        memory: "64Mi"
        cpu: "250m"
      limits:
        memory: "128Mi"
        cpu: "500m"`
      },
      cronjob: {
        label: 'CronJob',
        content: `apiVersion: batch/v1
kind: CronJob
metadata:
  name: my-cronjob
  namespace: default
spec:
  schedule: "*/1 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: my-cron-container
            image: busybox
            command: ["/bin/sh", "-c", "date; echo Hello from the Kubernetes cluster"]
          restartPolicy: OnFailure`
      }
    }
  },
  network: {
    label: 'Network',
    children: {
      service: {
        label: 'Service',
        content: `apiVersion: v1
kind: Service
metadata:
  name: my-service
  namespace: default
spec:
  selector:
    app: my-app
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
  type: ClusterIP`
      },
      ingress: {
        label: 'Ingress',
        content: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
  namespace: default
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: my-app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-service
            port:
              number: 80`
      }
    }
  },
  config: {
    label: 'Config',
    children: {
      configmap: {
        label: 'ConfigMap',
        content: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  namespace: default
data:
  config.yaml: |
    server:
      port: 8080
      host: 0.0.0.0
    database:
      host: localhost
      port: 5432
      name: mydb`
      },
      secret: {
        label: 'Secret',
        content: `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: default
type: Opaque
data:
  username: BASE64_ENCODED_USERNAME
  password: BASE64_ENCODED_PASSWORD
`
      }
    }
  },
  storage: {
    label: 'Storage',
    children: {
      pv: {
        label: 'PersistentVolume',
        content: `apiVersion: v1
kind: PersistentVolume
metadata:
  name: my-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: manual
  hostPath:
    path: /data/my-pv`
      },
      pvc: {
        label: 'PersistentVolumeClaim',
        content: `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: manual`
      }
    }
  }
};

interface BuiltinTemplateButtonProps {
  onSelectTemplate: (content: string) => void;
  style?: React.CSSProperties;
}

const BuiltinTemplateButton: React.FC<BuiltinTemplateButtonProps> = ({ onSelectTemplate, style }) => {
  const [visible, setVisible] = useState(false);

  const handleMenuClick = ({ key }: { key: string }) => {
    const keys = key.split('.');
    let current: any = BUILTIN_TEMPLATES;

    for (const k of keys) {
      if (current.children && current.children[k]) {
        current = current.children[k];
      } else if (current[k]) {
        current = current[k];
      } else {
        return;
      }
    }

    if (current.content) {
      onSelectTemplate(current.content);
    }
    setVisible(false);
  };

  const handleVisibleChange = (flag: boolean) => {
    setVisible(flag);
  };

  const buildMenuItems = (templates: any, parentKey = ''): any[] => {
    return Object.entries(templates).map(([key, value]: [string, any]) => {
      const fullKey = parentKey ? `${parentKey}.${key}` : key;

      if (value.children) {
        return {
          key: fullKey,
          label: value.label,
          children: buildMenuItems(value.children, fullKey)
        };
      } else if (value.content) {
        return {
          key: fullKey,
          label: value.label
        };
      }
      return null;
    }).filter(Boolean);
  };

  const menuItems = buildMenuItems(BUILTIN_TEMPLATES);

  const menu = {
    items: menuItems,
    onClick: handleMenuClick
  };

  return (
    <Dropdown
      menu={menu}
      onOpenChange={handleVisibleChange}
      open={visible}
      trigger={['click']}
    >
      <Button style={style}>
        内置模板 <DownOutlined />
      </Button>
    </Dropdown>
  );
};

export default BuiltinTemplateButton;
