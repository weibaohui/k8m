import React, { useState, useEffect } from 'react';
import { Button, List, Modal, Space, message } from 'antd';
import { fetcher } from '@/components/Amis/fetcher';
import JSZip from 'jszip';
import { saveAs } from 'file-saver';

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

const TemplatePanel: React.FC<TemplatePanelProps> = ({ onSelectTemplate, refreshKey = 0 }) => {
    const [templates, setTemplates] = useState<TemplateItem[]>([]);

    const [currentPage, setCurrentPage] = useState(1);
    const [total, setTotal] = useState(0);
    const pageSize = 10;


    const [selectedKind, setSelectedKind] = useState<string>('');
    const [resourceTypesList, setResourceTypesList] = useState<string[]>([]);
    useEffect(() => {
        const fetchTemplates = async () => {
            try {
                const response = await fetcher({
                    url: `/mgm/plugins/yaml_editor/template/list?page=${currentPage}&perPage=${pageSize}${selectedKind ? `&kind=${selectedKind}` : ''}`,
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
                    url: '/mgm/plugins/yaml_editor/template/kind/list',
                    method: 'get'
                });
                const data = response.data;
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




    const filteredTemplates = templates.filter(template =>
        !selectedKind || template.kind === selectedKind
    );

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
            </div>
        </List.Item>
    );

    return (
        <div>
            <div style={{
                border: '1px solid #d9d9d9',
                borderRadius: '4px',
                padding: '16px',
                position: 'relative',
                marginBottom: '10px'
            }}>
                <div style={{
                    position: 'absolute',
                    top: '-12px',
                    left: '12px',
                    background: '#FFFFFF',
                    padding: '0 8px',
                    fontWeight: '500'
                }}>
                    自定义模板
                </div>
                <div style={{ display: 'flex', gap: '8px' }}>
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
                                                        url: '/mgm/plugins/yaml_editor/template/save',
                                                        method: 'post',
                                                        data: template
                                                    });
                                                } catch (error) {
                                                    console.error('Failed to save template:', error);
                                                }
                                            }
                                            message.success(`成功导入 ${newTemplates.length} 个模板`);
                                            const response = await fetcher({
                                                url: `/mgm/plugins/yaml_editor/template/list?page=${currentPage}&perPage=${pageSize}`,
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
                        导入
                    </Button>
                    <Button
                        variant="outlined"
                        onClick={async () => {
                            try {
                                const firstPageResponse = await fetcher({
                                    url: `/mgm/plugins/yaml_editor/template/list?page=1&perPage=${pageSize}`,
                                    method: 'get'
                                });

                                if (firstPageResponse.data?.status !== 0) {
                                    throw new Error('获取模板数据失败');
                                }

                                //@ts-ignore
                                const totalCount = firstPageResponse.data.data.count;
                                const totalPages = Math.ceil(totalCount / pageSize);
                                let allTemplates: TemplateItem[] = [];

                                for (let page = 1; page <= totalPages; page++) {
                                    const response = await fetcher({
                                        url: `/mgm/plugins/yaml_editor/template/list?page=${page}&perPage=${pageSize}`,
                                        method: 'get'
                                    });

                                    //@ts-ignore
                                    if (response.data?.status === 0 && response.data?.data?.rows) {
                                        //@ts-ignore
                                        allTemplates.push(...response.data.data.rows);
                                    }
                                }

                                const zip = new JSZip();

                                const templatesByKind: { [key: string]: TemplateItem[] } = {};
                                allTemplates.forEach(template => {
                                    const kind = template.kind || '未分类';
                                    if (!templatesByKind[kind]) {
                                        templatesByKind[kind] = [];
                                    }
                                    templatesByKind[kind].push(template);
                                });

                                Object.entries(templatesByKind).forEach(([kind, templates]) => {
                                    templates.forEach(template => {
                                        const fileName = `${kind}/${template.name}.yaml`;
                                        zip.file(fileName, template.content);
                                    });
                                });

                                const blob = await zip.generateAsync({ type: 'blob' });
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
                        导出
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
                            fetcher({
                                url: '/mgm/plugins/yaml_editor/template/save',
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
                        新建
                    </Button>
                </div>
            </div>
            <div style={{ marginBottom: '10px', display: 'flex', gap: '8px' }}>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px', alignItems: 'center' }}>
                    <span style={{ fontSize: '14px', color: '#595959' }}>资源分类：</span>
                    <span
                        style={{
                            cursor: 'pointer',
                            padding: '4px 8px',
                            borderRadius: '4px',
                            backgroundColor: selectedKind === '' ? '#1890ff' : '#f0f0f0',
                            color: selectedKind === '' ? '#fff' : '#262626',
                            fontSize: '14px'
                        }}
                        onClick={() => {
                            setCurrentPage(1);
                            setSelectedKind('');
                        }}
                    >
                        全部
                    </span>
                    {resourceTypesList.map(type => (
                        <span
                            key={type}
                            style={{
                                cursor: 'pointer',
                                padding: '4px 8px',
                                borderRadius: '4px',
                                backgroundColor: selectedKind === type ? '#1890ff' : '#f0f0f0',
                                color: selectedKind === type ? '#fff' : '#262626',
                                fontSize: '14px'
                            }}
                            onClick={() => {
                                setCurrentPage(1);
                                setSelectedKind(type);
                            }}
                        >
                            {type}
                        </span>
                    ))}
                </div>
            </div>
            <List
                dataSource={filteredTemplates}
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
        </div>
    );
};

export default TemplatePanel;
