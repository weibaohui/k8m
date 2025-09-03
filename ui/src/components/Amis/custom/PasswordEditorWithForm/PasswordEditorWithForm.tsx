import React, { useState } from 'react';

import { replacePlaceholders } from "@/utils/utils.ts";
import { fetcher } from "@/components/Amis/fetcher.ts";
import { Button, Input, message } from 'antd';
import { encrypt } from "@/utils/crypto.ts";


interface PasswordEditorWithFormProps {
    api: string;
    data: Record<string, any>
}

const PasswordEditorWithForm: React.FC<PasswordEditorWithFormProps> = ({
    api,
    data,
}) => {
    const [loading, setLoading] = useState(false);
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');

    if (api) {
        api = replacePlaceholders(api, data)
    }

    /**
     * 处理密码保存逻辑
     * 验证两次密码输入是否一致，一致才允许提交
     */
    const handleSave = async () => {
        if (!api) return;
        if (!password) {
            message.error('请输入密码');
            return;
        }
        if (!confirmPassword) {
            message.error('请输入确认密码');
            return;
        }
        if (password !== confirmPassword) {
            message.error('两次输入的密码不一致，请重新输入');
            return;
        }
        setLoading(true);

        const encryptedPassword = encrypt(password);
        const encryptedConfirmPassword = encrypt(confirmPassword);

        const response = await fetcher({
            url: api,
            method: 'post',
            data: {
                password: encryptedPassword,
                confirmPassword: encryptedConfirmPassword
            }
        });

        if (response.data?.status !== 0) {
            message.error(`密码修改失败:请尝试刷新后重试。 ${response.data?.msg}`);
        } else {
            message.info('密码修改成功！');
            setPassword('');
            setConfirmPassword('');
        }
        setLoading(false);
    };

    return (
        <>
            <div style={{ width: '100%', height: 'calc(100vh - 200px)', display: 'flex', flexDirection: 'column' }}>
                <div style={{ padding: '10px', display: 'flex', flexDirection: 'column', gap: '10px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                        <span style={{ minWidth: '80px' }}>新密码:</span>
                        <Input.Password
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="请输入新密码"
                            style={{ flex: 1 }}
                        />
                    </div>
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
                    {confirmPassword && password !== confirmPassword && (
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
                                disabled={!password || !confirmPassword || password !== confirmPassword}
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
