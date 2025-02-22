import React, {useRef, useState, useEffect} from 'react';
import {Button, List, Input, Modal, Space, Drawer, Select, InputRef, Divider, message} from 'antd';
import {DeleteFilled, EditFilled, PlusOutlined} from '@ant-design/icons';
import {fetcher} from '@/components/Amis/fetcher';
import * as monaco from 'monaco-editor';
import JSZip from 'jszip';
import {saveAs} from 'file-saver';

interface TemplateItem {
    id?: string;
    name: string;
    kind: string;
    content: string;
}

interface TemplatePanelProps {
    onSelectTemplate: (content: string) => void;
    refreshKey?: number;
}

const TemplatePanel: React.FC<TemplatePanelProps> = ({onSelectTemplate, refreshKey = 0}) => {
    const [templates, setTemplates] = useState<TemplateItem[]>([]);

    const [currentPage, setCurrentPage] = useState(1);
    const [total, setTotal] = useState(0);
    const pageSize = 10;


    const [editingTemplate, setEditingTemplate] = useState<TemplateItem | null>(null);
    const [drawerVisible, setDrawerVisible] = useState(false);
    const [selectedKind, setSelectedKind] = useState<string>('');
    const [editForm, setEditForm] = useState({
        name: '',
        kind: '',
        content: ''
    });
    const [newKind, setNewKind] = useState('');
    const [resourceTypesList, setResourceTypesList] = useState<string[]>([]);
    useEffect(() => {
        const fetchTemplates = async () => {
            try {
                const response = await fetcher({
                    url: `/mgm/custom/template/list?page=${currentPage}&perPage=${pageSize}${selectedKind ? `&kind=${selectedKind}` : ''}`,
                    method: 'get'
                });
                const data = response.data;
                //@ts-ignore
                if (data?.status === 0 && data?.data?.rows) {
                    //@ts-ignore
                    setTemplates(data.data.rows);
                    //@ts-ignore
                    setTotal(data.data.count || 0);
                }
            } catch (error) {
                console.error('Failed to fetch templates:', error);
                Modal.error({
                    title: '获取模板列表失败',
                    content: '无法从服务器获取模板列表'
                });
            }
        };
        fetchTemplates();
        const fetchResourceTypes = async () => {
            try {
                const response = await fetcher({
                    url: '/mgm/custom/template/kind/list',
                    method: 'get'
                });
                const data = await response.data;
                //@ts-ignore
                if (data?.data?.rows) {
                    //@ts-ignore
                    const types = data.data.rows.map(item => item.kind);
                    setResourceTypesList(types);
                }
            } catch (error) {
                console.error('Failed to fetch resource types:', error);
                Modal.error({
                    title: '获取资源类型失败',
                    content: '无法从服务器获取资源类型列表'
                });
            }
        };
        fetchResourceTypes();
    }, [currentPage, selectedKind, refreshKey]);

    const inputRef = useRef<InputRef>(null);


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

    const onNewKindChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const value = event.target.value;
        setNewKind(value);
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

    const handleEditSubmit = async () => {
        if (editingTemplate && editForm.name.trim()) {
            try {
                const response = await fetcher({
                    url: '/mgm/custom/template/save',
                    method: 'post',
                    data: {
                        id: editingTemplate.id,
                        ...editForm
                    }
                });

                if (response.data?.status === 0) {
                    setTemplates(prevTemplates =>
                        prevTemplates.map(template =>
                            template.id === editingTemplate.id
                                ? {...template, ...editForm}
                                : template
                        )
                    );
                    setDrawerVisible(false);
                    setEditingTemplate(null);
                    message.success('模板已成功更新');
                } else {
                    throw new Error(response.data?.msg || '更新失败');
                }
            } catch (error) {
                console.error('Failed to update template:', error);
                Modal.error({
                    title: '保存失败',
                    content: '无法更新模板：' + (error instanceof Error ? error.message : '未知错误')
                });
            }
        }
    };

    const handleDelete = (templateId?: string) => {
        if (!templateId) {
            return;
        }
        Modal.confirm({
            title: '确认删除',
            content: '确定要删除这个模板吗？',
            onOk: async () => {
                try {
                    const response = await fetcher({
                        url: `/mgm/custom/template/delete/${templateId}`,
                        method: 'delete'
                    });

                    if (response.data?.status === 0) {
                        setTemplates(prevTemplates => prevTemplates.filter(t => t.id !== templateId));
                        message.success('模板已成功删除');
                    } else {
                        throw new Error(response.data?.msg || '删除失败');
                    }
                } catch (error) {
                    console.error('Failed to delete template:', error);
                    Modal.error({
                        title: '删除失败',
                        content: '无法删除模板：' + (error instanceof Error ? error.message : '未知错误')
                    });
                }
            }
        });
    };

    const renderTemplate = (template: TemplateItem) => (
        <List.Item key={template.id} className="list-item" style={{cursor: 'pointer'}}
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
                <div style={{display: 'flex', gap: '8px', zIndex: 10}}>
                    <Button
                        type="text"
                        icon={<EditFilled style={{color: '#1890ff'}}/>}
                        onClick={(e) => {
                            e.stopPropagation();
                            handleNameEdit(template);
                        }}
                    />
                    <Button
                        type="text"
                        icon={<DeleteFilled style={{color: '#f23034'}}/>}
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
            <div style={{marginBottom: '10px', display: 'flex', gap: '8px'}}>
                <Button
                    variant="outlined"
                    onClick={async () => {
                        const input = document.createElement('input');
                        input.type = 'file';
                        input.accept = '.zip';
                        input.onchange = async (e) => {
                            const file = (e.target as HTMLInputElement).files?.[0];
                            if (file) {
                                try {
                                    const zip = await JSZip.loadAsync(file);
                                    const newTemplates: TemplateItem[] = [];

                                    for (const [filePath, fileData] of Object.entries(zip.files)) {
                                        if (!fileData.dir && (filePath.endsWith('.yaml') || filePath.endsWith('.yml'))) {
                                            const content = await fileData.async('text');
                                            const pathParts = filePath.split('/');
                                            const kind = pathParts.length > 1 ? pathParts[0] : '';
                                            const name = pathParts[pathParts.length - 1].replace(/\.(yaml|yml)$/, '');

                                            newTemplates.push({
                                                content,
                                                kind,
                                                name
                                            });
                                        }
                                    }

                                    if (newTemplates.length > 0) {
                                        for (const template of newTemplates) {
                                            try {
                                                await fetcher({
                                                    url: '/mgm/custom/template/save',
                                                    method: 'post',
                                                    data: template
                                                });
                                            } catch (error) {
                                                console.error('Failed to save template:', error);
                                            }
                                        }
                                        message.success(`成功导入 ${newTemplates.length} 个模板`);
                                        // 刷新模板列表
                                        const response = await fetcher({
                                            url: `/mgm/custom/template/list?page=${currentPage}&perPage=${pageSize}`,
                                            method: 'get'
                                        });
                                        //@ts-ignore
                                        if (response.data?.status === 0 && response.data?.data?.rows) {
                                            //@ts-ignore
                                            setTemplates(response.data.data.rows);
                                            //@ts-ignore
                                            setTotal(response.data.data.count || 0);
                                        }
                                    } else {
                                        Modal.warning({
                                            title: '导入提示',
                                            content: 'ZIP文件中没有找到YAML文件'
                                        });
                                    }
                                } catch (error) {
                                    Modal.error({
                                        title: '导入失败',
                                        content: '无法解析ZIP文件或文件格式错误'
                                    });
                                }
                            }
                        };
                        input.click();
                    }}
                >
                    导入模板
                </Button>
                <Button
                    variant="outlined"
                    onClick={async () => {
                        try {
                            // 获取第一页数据以获取总数
                            const firstPageResponse = await fetcher({
                                url: `/mgm/custom/template/list?page=1&perPage=${pageSize}`,
                                method: 'get'
                            });

                            if (firstPageResponse.data?.status !== 0) {
                                throw new Error('获取模板数据失败');
                            }

                            //@ts-ignore
                            const totalCount = firstPageResponse.data.data.count;
                            const totalPages = Math.ceil(totalCount / pageSize);
                            let allTemplates: TemplateItem[] = [];

                            // 循环获取所有页面的数据
                            for (let page = 1; page <= totalPages; page++) {
                                const response = await fetcher({
                                    url: `/mgm/custom/template/list?page=${page}&perPage=${pageSize}`,
                                    method: 'get'
                                });

                                //@ts-ignore
                                if (response.data?.status === 0 && response.data?.data?.rows) {
                                    //@ts-ignore
                                    allTemplates.push(...response.data.data.rows);
                                }
                            }

                            const zip = new JSZip();

                            // 按kind分组创建文件夹
                            const templatesByKind: { [key: string]: TemplateItem[] } = {};
                            allTemplates.forEach(template => {
                                const kind = template.kind || '未分类';
                                if (!templatesByKind[kind]) {
                                    templatesByKind[kind] = [];
                                }
                                templatesByKind[kind].push(template);
                            });

                            // 创建文件夹结构并添加文件
                            Object.entries(templatesByKind).forEach(([kind, templates]) => {
                                templates.forEach(template => {
                                    const fileName = `${kind}/${template.name}.yaml`;
                                    zip.file(fileName, template.content);
                                });
                            });

                            const blob = await zip.generateAsync({type: 'blob'});
                            saveAs(blob, 'templates.zip');
                            message.success(`成功导出 ${allTemplates.length} 个模板`);
                            allTemplates = []
                        } catch (error) {
                            console.error('导出模板失败:', error);
                            Modal.error({
                                title: '导出失败',
                                content: '导出模板时发生错误：' + (error instanceof Error ? error.message : '未知错误')
                            });
                        }
                    }}
                >
                    导出模板
                </Button>
                <Button
                    variant="outlined"
                    onClick={() => {
                        //@ts-ignore
                        const newTemplate: TemplateItem = {
                            name: `模板 ${templates.length + 1}`,
                            content: '',
                            kind: selectedKind
                        };
                        // 调用后端API保存新模板
                        fetcher({
                            url: '/mgm/custom/template/save',
                            method: 'post',
                            data: newTemplate
                        }).then(response => {
                            if (response.data?.status === 0) {
                                const savedTemplate = {
                                    ...newTemplate,
                                    //@ts-ignore
                                    id: response.data.data.id || Math.random().toString(36).substring(2, 15)
                                };
                                setTemplates(prev => [...prev, savedTemplate]);
                                message.success('新模板已成功创建');
                            } else {
                                throw new Error(response.data?.msg || '创建失败');
                            }
                        }).catch(error => {
                            console.error('Failed to create template:', error);
                            Modal.error({
                                title: '创建失败',
                                content: '无法创建新模板：' + error.message
                            });
                        });
                    }}
                >
                    新建模板
                </Button>
            </div>
            <div style={{marginBottom: '10px', display: 'flex', gap: '8px'}}>
                <Select
                    style={{width: 200}}
                    value={selectedKind}
                    onChange={(value) => {
                        setCurrentPage(1);
                        setSelectedKind(value);
                    }}
                    placeholder="按资源分类筛选"
                    allowClear
                    options={resourceTypesList.map(type => ({label: type, value: type}))}
                />
            </div>
            <List
                dataSource={filteredTemplates}
                renderItem={renderTemplate}
                bordered={true}
            />
            <div style={{marginTop: '16px', textAlign: 'right'}}>
                <Space.Compact>
                    <Button
                        type="default"
                        disabled={currentPage === 1}
                        onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                    >
                        上一页
                    </Button>
                    <Button type="default" disabled>
                        {currentPage}/{Math.ceil(total / pageSize)}
                    </Button>
                    <Button
                        type="default"
                        disabled={currentPage >= Math.ceil(total / pageSize)}
                        onClick={() => setCurrentPage(prev => Math.min(Math.ceil(total / pageSize), prev + 1))}
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
                    <div style={{textAlign: 'right'}}>
                        <Space>
                            <Button onClick={() => setDrawerVisible(false)}>取消</Button>
                            <Button type="primary" onClick={handleEditSubmit}>保存</Button>
                        </Space>
                    </div>
                }
            >
                <div style={{display: 'flex', flexDirection: 'column', gap: '16px'}}>
                    <div>
                        <div style={{marginBottom: '8px'}}>模板名称</div>
                        <Input
                            value={editForm.name}
                            onChange={(e) => setEditForm(prev => ({...prev, name: e.target.value}))}
                            placeholder="请输入模板名称"
                        />
                    </div>
                    <div>
                        <div style={{marginBottom: '8px'}}>资源分类</div>
                        <Select
                            value={editForm.kind}
                            onChange={(value) => setEditForm(prev => ({...prev, kind: value}))}
                            placeholder="请选择资源分类"
                            options={resourceTypesList.map(type => ({label: type, value: type}))}
                            style={{width: 200}}
                            allowClear
                            dropdownRender={(menu) => (
                                <>
                                    {menu}
                                    <Divider style={{margin: '8px 0'}}/>
                                    <Space style={{padding: '0 8px 4px'}}>
                                        <Input
                                            placeholder="请输入新的资源分类"
                                            ref={inputRef}
                                            value={newKind}
                                            onChange={onNewKindChange}
                                            onKeyDown={(e) => e.stopPropagation()}
                                        />
                                        <Button type="text" icon={<PlusOutlined/>} onClick={handleAddKind}>
                                            添加分类
                                        </Button>
                                    </Space>
                                </>
                            )}
                        />
                    </div>
                    <div>
                        <div style={{marginBottom: '8px'}}>模板内容</div>
                        <div
                            ref={(el) => {
                                if (el && !el.hasChildNodes()) {
                                    const editor = monaco.editor.create(el, {
                                        value: editForm.content,
                                        language: 'yaml',
                                        theme: 'vs',
                                        automaticLayout: true,
                                        minimap: {
                                            enabled: false
                                        },
                                        scrollbar: {
                                            vertical: 'auto'
                                        },
                                        wordWrap: 'on'
                                    });
                                    editor.onDidChangeModelContent(() => {
                                        setEditForm(prev => ({
                                            ...prev,
                                            content: editor.getValue()
                                        }));
                                    });
                                }
                            }}
                            style={{height: '400px', border: '1px solid #d9d9d9'}}
                        />
                    </div>
                </div>
            </Drawer>
        </div>
    );
};

export default TemplatePanel;