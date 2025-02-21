import React, { useRef, useState } from 'react';
import { Button, List, Input, Modal, Space, Drawer, Select, InputRef, Divider } from 'antd';
import { DeleteFilled, EditFilled, PlusOutlined } from '@ant-design/icons';

interface TemplateItem {
    id: string;
    name: string;
    kind: string;
    content: string;
}

interface TemplatePanelProps {
    onSelectTemplate: (content: string) => void;
}

const defaultTemplates: TemplateItem[] = [
    {
        id: '1',
        name: 'Nginx Deployment',
        kind: 'Deployment',
        content: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80`
    },
    {
        id: '2',
        name: 'Redis Service',
        kind: 'Service',
        content: `apiVersion: v1
kind: Service
metadata:
  name: redis-service
spec:
  selector:
    app: redis
  ports:
    - protocol: TCP
      port: 6379
      targetPort: 6379`
    },
    {
        id: '3',
        name: 'MySQL ConfigMap',
        kind: 'ConfigMap',
        content: `apiVersion: v1
kind: ConfigMap
metadata:
  name: mysql-config
data:
  mysql.conf: |
    [mysqld]
    max_connections=250
    character-set-server=utf8mb4
    collation-server=utf8mb4_unicode_ci`
    },
    {
        id: '4',
        name: 'MongoDB StatefulSet',
        kind: 'StatefulSet',
        content: `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mongodb
spec:
  serviceName: mongodb
  replicas: 3
  selector:
    matchLabels:
      app: mongodb
  template:
    metadata:
      labels:
        app: mongodb
    spec:
      containers:
      - name: mongodb
        image: mongo:4.4`
    },
    {
        id: '5',
        name: 'Persistent Volume Claim',
        kind: 'PersistentVolumeClaim',
        content: `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mongodb-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi`
    }
];

const TemplatePanel: React.FC<TemplatePanelProps> = ({ onSelectTemplate }) => {
    const [templates, setTemplates] = useState<TemplateItem[]>(defaultTemplates);
    const [currentPage, setCurrentPage] = useState(1);
    const [editingTemplate, setEditingTemplate] = useState<TemplateItem | null>(null);
    const [drawerVisible, setDrawerVisible] = useState(false);
    const [selectedKind, setSelectedKind] = useState<string>('');
    const [editForm, setEditForm] = useState({
        name: '',
        kind: '',
        content: ''
    });
    const [newKind, setNewKind] = useState('');
    const [resourceTypesList, setResourceTypesList] = useState([
        'Deployment',
        'Service',
        'ConfigMap',
        'StatefulSet',
        'DaemonSet',
        'Job',
        'CronJob',
        'PersistentVolumeClaim',
        'Secret',
        'Ingress',
        'NetworkPolicy'
    ]);
    const inputRef = useRef<InputRef>(null);

    const pageSize = 10;

    const handleAddKind = (e: React.MouseEvent<HTMLButtonElement | HTMLAnchorElement>) => {
        e.preventDefault();
        if (newKind && !resourceTypesList.includes(newKind)) {
            setResourceTypesList([...resourceTypesList, newKind]);
            setNewKind('');
            setTimeout(() => {
                inputRef.current?.focus();
            }, 0);
        }
    };

    const onNewKindChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setNewKind(event.target.value);
    };

    const filteredTemplates = templates.filter(template =>
        !selectedKind || template.kind === selectedKind
    );
    const handleNameEdit = (template: TemplateItem) => {
        setEditingTemplate(template);
        setEditForm({
            name: template.name,
            kind: template.kind,
            content: template.content
        });
        setDrawerVisible(true);
    };

    const handleEditSubmit = () => {
        if (editingTemplate && editForm.name.trim()) {
            setTemplates(prevTemplates =>
                prevTemplates.map(template =>
                    template.id === editingTemplate.id
                        ? { ...template, ...editForm }
                        : template
                )
            );
            setDrawerVisible(false);
            setEditingTemplate(null);
        }
    };

    const handleDelete = (templateId: string) => {
        Modal.confirm({
            title: '确认删除',
            content: '确定要删除这个模板吗？',
            onOk: () => {
                setTemplates(prevTemplates => prevTemplates.filter(t => t.id !== templateId));
            }
        });
    };

    const renderTemplate = (template: TemplateItem) => (
        <List.Item key={template.id} className="list-item" style={{ cursor: 'pointer' }}
            onClick={() => onSelectTemplate(template.content)}>
            <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                width: '100%',
                position: 'relative',
                backgroundColor: '#FFFFFF'
            }}>
                <div style={{
                    flex: 1,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                    marginRight: '10px'
                }}>
                    {template.name}
                </div>
                <div style={{ display: 'flex', gap: '8px', zIndex: 10 }}>
                    <Button
                        type="text"
                        icon={<EditFilled style={{ color: '#1890ff' }} />}
                        onClick={(e) => {
                            e.stopPropagation();
                            handleNameEdit(template);
                        }}
                    />
                    <Button
                        type="text"
                        icon={<DeleteFilled style={{ color: '#f23034' }} />}
                        onClick={(e) => {
                            e.stopPropagation();
                            handleDelete(template.id);
                        }}
                    />
                </div>
            </div>
        </List.Item>
    );

    return (
        <div>
            <div style={{ marginBottom: '10px', display: 'flex', gap: '8px' }}>
                <Space.Compact>
                    <Button
                        variant="outlined"
                        onClick={() => {
                            const newTemplate: TemplateItem = {
                                id: Math.random().toString(36).substring(2, 15),
                                name: `模板 ${templates.length + 1}`,
                                content: '',
                                kind: selectedKind
                            };
                            setTemplates(prev => [...prev, newTemplate]);
                        }}
                    >
                        新建模板
                    </Button>
                </Space.Compact>
                <Select
                    style={{ width: 200 }}
                    value={selectedKind}
                    onChange={(value) => {
                        setSelectedKind(value);
                        setCurrentPage(1);
                    }}
                    placeholder="按资源分类筛选"
                    allowClear
                    dropdownRender={(menu) => (
                        <>
                            {menu}
                            <Divider style={{ margin: '8px 0' }} />
                            <Space style={{ padding: '0 8px 4px' }}>
                                <Input
                                    placeholder="请输入新的资源分类"
                                    ref={inputRef}
                                    value={newKind}
                                    onChange={onNewKindChange}
                                    onKeyDown={(e) => e.stopPropagation()}
                                />
                                <Button type="text" icon={<PlusOutlined />} onClick={handleAddKind}>
                                    添加类型
                                </Button>
                            </Space>
                        </>
                    )}
                    options={resourceTypesList.map(type => ({ label: type, value: type }))}
                />
            </div>
            <List
                dataSource={filteredTemplates.slice((currentPage - 1) * pageSize, currentPage * pageSize)}
                renderItem={renderTemplate}
                bordered={true}
            />
            <div style={{ marginTop: '16px', textAlign: 'right' }}>
                <Space.Compact>
                    <Button
                        type="default"
                        disabled={currentPage === 1}
                        onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                    >
                        上一页
                    </Button>
                    <Button type="default" disabled>
                        {currentPage}/{Math.ceil(templates.length / pageSize)}
                    </Button>
                    <Button
                        type="default"
                        disabled={currentPage >= Math.ceil(templates.length / pageSize)}
                        onClick={() => setCurrentPage(prev => Math.min(Math.ceil(templates.length / pageSize), prev + 1))}
                    >
                        下一页
                    </Button>
                </Space.Compact>
            </div>

            <Drawer
                title="编辑模板"
                width={600}
                open={drawerVisible}
                onClose={() => setDrawerVisible(false)}
                footer={
                    <div style={{ textAlign: 'right' }}>
                        <Space>
                            <Button onClick={() => setDrawerVisible(false)}>取消</Button>
                            <Button type="primary" onClick={handleEditSubmit}>保存</Button>
                        </Space>
                    </div>
                }
            >
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    <div>
                        <div style={{ marginBottom: '8px' }}>模板名称</div>
                        <Input
                            value={editForm.name}
                            onChange={(e) => setEditForm(prev => ({ ...prev, name: e.target.value }))}
                            placeholder="请输入模板名称"
                        />
                    </div>
                    <div>
                        <div style={{ marginBottom: '8px' }}>资源分类</div>
                        <Select
                            value={editForm.kind}
                            onChange={(value) => setEditForm(prev => ({ ...prev, kind: value }))}
                            placeholder="请选择资源分类"
                            options={resourceTypesList.map(type => ({ label: type, value: type }))}
                        />
                    </div>
                    <div>
                        <div style={{ marginBottom: '8px' }}>模板内容</div>
                        <Input.TextArea
                            value={editForm.content}
                            onChange={(e) => setEditForm(prev => ({ ...prev, content: e.target.value }))}
                            placeholder="请输入YAML内容"
                            rows={15}
                        />
                    </div>
                </div>
            </Drawer>
        </div>
    );
};

export default TemplatePanel;