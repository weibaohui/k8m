import React, { useState } from 'react';
import { Button, Dropdown, Menu } from 'antd';
import { DownOutlined } from '@ant-design/icons';

// 内置模板数据
const BUILTIN_TEMPLATES = [
    {
        key: 'deployment',
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
    {
        key: 'service',
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
    {
        key: 'configmap',
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
    {
        key: 'secret',
        label: 'Secret',
        content: `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: default
type: Opaque
data:
  username: YWRtaW4=  # admin (base64 encoded)
  password: MWYyZDFlMmU2N2Rm  # 1f2d1e2e67df (base64 encoded)`
    },
    {
        key: 'ingress',
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
    },
    {
        key: 'statefulset',
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
    }
];

interface BuiltinTemplateButtonProps {
    onSelectTemplate: (content: string) => void;
    style?: React.CSSProperties;
}

/**
 * 内置模板按钮组件
 * 提供常用的Kubernetes资源模板选择功能
 */
const BuiltinTemplateButton: React.FC<BuiltinTemplateButtonProps> = ({ onSelectTemplate, style }) => {
    const [visible, setVisible] = useState(false);

    /**
     * 处理模板选择
     * @param key 模板键值
     */
    const handleMenuClick = ({ key }: { key: string }) => {
        const template = BUILTIN_TEMPLATES.find(t => t.key === key);
        if (template) {
            onSelectTemplate(template.content);
        }
        setVisible(false);
    };

    /**
     * 处理下拉菜单显示状态变化
     * @param flag 显示状态
     */
    const handleVisibleChange = (flag: boolean) => {
        setVisible(flag);
    };

    // 构建菜单项
    const menuItems = BUILTIN_TEMPLATES.map(template => ({
        key: template.key,
        label: template.label
    }));

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