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

    if (api) {
        api = replacePlaceholders(api, data)
    }

    const handleSave = async () => {
        if (!api) return;
        if (!password) {
            message.error('请输入密码');
            return;
        }
        setLoading(true);

        const encryptedPassword = encrypt(password);

        const response = await fetcher({
            url: api,
            method: 'post',
            data: {
                password: encryptedPassword
            }
        });

        if (response.data?.status !== 0) {
            message.error(`密码修改失败:请尝试刷新后重试。 ${response.data?.msg}`);
        } else {
            message.info('密码修改成功！');
            setPassword('');
        }
        setLoading(false);
    };

    return (
        <>
            <div style={{ width: '100%', height: 'calc(100vh - 200px)', display: 'flex', flexDirection: 'column' }}>
                <div style={{ padding: '10px', display: 'flex', justifyContent: 'flex-end' }}>
                    <Input.Password
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        placeholder="请输入密码"
                    />
                    {api && <Button type="primary" onClick={handleSave} loading={loading}>保存</Button>}
                </div>
            </div>
        </>
    );
};

export default PasswordEditorWithForm;
