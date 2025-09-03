import React, { useState } from 'react';

import { replacePlaceholders } from "@/utils/utils.ts";
import { fetcher } from "@/components/Amis/fetcher.ts";
import { Button, Input, message } from 'antd';
import { encrypt } from "@/utils/crypto.ts";


interface PasswordEditorWithFormProps {
    api: string;
    data: Record<string, any>;
    isAdmin?: boolean; // 是否为管理员模式，true表示管理员可直接修改密码，false表示普通用户需要验证原密码
}

const PasswordEditorWithForm: React.FC<PasswordEditorWithFormProps> = ({
    api,
    data,
    isAdmin = false, // 默认为普通用户模式
}) => {
    const [oldPassword, setOldPassword] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [loading, setLoading] = useState(false);

    if (api) {
        api = replacePlaceholders(api, data)
    }

    /**
     * 处理密码保存逻辑
     * 根据isAdmin参数决定验证逻辑：管理员模式只需验证新密码，普通用户需要验证原密码
     */
    const handleSave = async () => {
        // 普通用户模式需要验证原密码
        if (!isAdmin && !oldPassword) {
            message.error('请输入原密码');
            return;
        }
        if (!password) {
            message.error('请输入新密码');
            return;
        }
        // 管理员模式下不需要确认密码
        if (!isAdmin && !confirmPassword) {
            message.error('请输入确认密码');
            return;
        }
        // 普通用户模式需要验证两次密码一致性
        if (!isAdmin && password !== confirmPassword) {
            message.error('两次输入的密码不一致，请重新输入');
            return;
        }
        setLoading(true);

        const encryptedPassword = encrypt(password);
        const requestData: any = {
            password: encryptedPassword
        };

        // 普通用户模式需要发送原密码和确认密码
        if (!isAdmin) {
            requestData.oldPassword = encrypt(oldPassword);
            requestData.confirmPassword = encrypt(confirmPassword);
        }

        const response = await fetcher({
            url: api,
            method: 'post',
            data: requestData
        });

        if (response.data?.status !== 0) {
            message.error(`密码修改失败:请尝试刷新后重试。 ${response.data?.msg}`);
        } else {
            message.info('密码修改成功！');
            setOldPassword('');
            setPassword('');
            setConfirmPassword('');
        }
        setLoading(false);
    };

    return (
        <>
            <div style={{ width: '100%', height: 'calc(100vh - 200px)', display: 'flex', flexDirection: 'column' }}>
                <div style={{ padding: '10px', display: 'flex', flexDirection: 'column', gap: '10px' }}>
                    {/* 普通用户模式显示原密码输入框 */}
                    {!isAdmin && (
                        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                            <span style={{ minWidth: '80px' }}>原密码:</span>
                            <Input.Password
                                value={oldPassword}
                                onChange={(e) => setOldPassword(e.target.value)}
                                placeholder="请输入原密码"
                                style={{ flex: 1 }}
                            />
                        </div>
                    )}
                    <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                        <span style={{ minWidth: '80px' }}>{isAdmin ? '密码:' : '新密码:'}</span>
                        <Input.Password
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder={isAdmin ? '请输入密码' : '请输入新密码'}
                            style={{ flex: 1 }}
                        />
                    </div>
                    {/* 普通用户模式显示确认密码输入框 */}
                    {!isAdmin && (
                        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                            <span style={{ minWidth: '80px' }}>确认密码:</span>
                            <Input.Password
                                value={confirmPassword}
                                onChange={(e) => setConfirmPassword(e.target.value)}
                                placeholder="请再次输入密码"
                                style={{ flex: 1 }}
                                status={confirmPassword && password !== confirmPassword ? 'error' : ''}
                            />
                        </div>
                    )}
                    {/* 普通用户模式显示密码不一致提示 */}
                    {!isAdmin && confirmPassword && password !== confirmPassword && (
                        <div style={{ color: '#ff4d4f', fontSize: '12px', marginLeft: '90px' }}>
                            两次输入的密码不一致
                        </div>
                    )}
                    <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '10px' }}>
                        {api && (
                            <Button 
                                type="primary" 
                                onClick={handleSave} 
                                loading={loading}
                                disabled={
                                    isAdmin 
                                        ? !password // 管理员模式只需要密码不为空
                                        : !oldPassword || !password || !confirmPassword || password !== confirmPassword // 普通用户模式需要三个字段都不为空且密码一致
                                }
                            >
                                保存
                            </Button>
                        )}
                    </div>
                </div>
            </div>
        </>
    );
};

export default PasswordEditorWithForm;
