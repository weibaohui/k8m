import { Form, Input, Button, Checkbox, message, Space } from 'antd'
import { useNavigate } from 'react-router-dom'
import {
    UserOutlined,
    LockOutlined,
    SafetyOutlined
} from '@ant-design/icons'
import styles from './index.module.scss'
import { useCallback, useEffect, useState } from 'react'
import { encrypt, decrypt } from '@/utils/crypto'

const FormItem = Form.Item

interface SSOConfig {
    name: string;
    type: string;
}

const Login = () => {
    const navigate = useNavigate()
    const [form] = Form.useForm();
    const [ssoConfigs, setSsoConfigs] = useState<SSOConfig[]>([]);
    const [loadingSSO, setLoadingSSO] = useState<Record<string, boolean>>({});
    const [isLdap, setIsLdap] = useState(false);
    const [ldapEnabled, setLdapEnabled] = useState(false);

    // 获取SSO配置
    useEffect(() => {
        fetch('/auth/sso/config')
            .then(res => res.json())
            .then(data => {
                if (data.status === 0 && Array.isArray(data.data)) {
                    setSsoConfigs(data.data);
                }
            })
            .catch(() => message.error('获取SSO配置失败'));
    }, []);

    // 获取LDAP配置
    useEffect(() => {
        fetch('/auth/ldap/config')
            .then(res => res.json())
            .then(data => {
                if (data.status === 0) {
                    setLdapEnabled(data.data.enabled);
                }
            })
            .catch(() => message.error('获取LDAP配置失败'));
    }, []);

    // useEffect 读取 remember 数据
    useEffect(() => {
        const savedData = localStorage.getItem('remember');
        if (savedData) {
            const parsedData = JSON.parse(savedData);
            form.setFieldsValue({
                username: parsedData.username,
            });

            // 解密密码
            if (parsedData.password) {
                const decryptedPassword = decrypt(parsedData.password);
                form.setFieldValue('password', decryptedPassword);
            }

            form.setFieldValue('remember', parsedData.remember === true);
        }
    }, [form]);

    const onSubmit = useCallback(() => {
        form.validateFields().then(async (values) => {
            try {
                const encryptedPassword = encrypt(values.password);
                const res = await fetch('/auth/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        username: values.username,
                        password: encryptedPassword,  // 发送加密后的密码
                        code: values.code, // 添加2FA验证码
                        loginType: isLdap ? 1 : 0 // 0: 普通登录, 1: LDAP登录
                    }),
                });
                const data = await res.json();
                if (res.ok) {
                    message.success('登录成功');
                    localStorage.setItem('token', data.token);
                    // 记住密码逻辑
                    const rememberData = {
                        username: values.username,
                        password: encryptedPassword,  // 存储加密后的密码
                        remember: values.remember,
                    };

                    if (values.remember) {
                        localStorage.setItem('remember', JSON.stringify(rememberData));  // 存储加密后的密码
                    } else {
                        localStorage.removeItem('remember');
                    }

                    navigate('/');
                } else {
                    message.error(data.message || '登录失败');
                }
            } catch (error) {
                message.error('网络错误');
            }
        });
    }, [navigate, form,isLdap]);

    return <section className={styles.login}>
        <div className={styles.content}>
            <Form
                form={form}
                onKeyDown={(event) => {
                    if (event.key === 'Enter') {
                        event.preventDefault();
                        onSubmit();
                    }
                }}
                className={styles.form}
                autoComplete='off'
            >
                <div>
                    <h2 style={{ color: '#666', fontSize: '24px', marginBottom: 20 }}>欢迎登录</h2>
                </div>
                <FormItem name='username' rules={[{ required: true, message: '请输入用户名' }]}>
                    <Input placeholder='请输入用户名' prefix={<UserOutlined />} />
                </FormItem>
                <FormItem name='password' rules={[{ required: true, message: '请输入密码' }]}>
                    <Input.Password
                        prefix={<LockOutlined />}
                        placeholder='请输入密码'
                    />
                </FormItem>
                <FormItem name='code'>
                    <Input
                        prefix={<SafetyOutlined />}
                        placeholder='请输入2FA验证码，未开启可不填'
                    />
                </FormItem>
                <div style={{ display: 'flex', justifyContent: 'flex-start', gap: '24px', alignItems: 'center' }}>
                    <FormItem name='remember' valuePropName='checked' style={{ marginBottom: 0 }}>
                        <Checkbox>记住</Checkbox>
                    </FormItem>
                    {ldapEnabled && (
                        <FormItem name='ldap' valuePropName='checked' style={{ marginBottom: 0 }}>
                            <Checkbox onChange={(e) => setIsLdap(e.target.checked)}>LDAP登录</Checkbox>
                        </FormItem>
                    )}
                </div>
                <FormItem>
                    <Button type='primary' block onClick={onSubmit}>登 录</Button>
                </FormItem>
                {ssoConfigs.length > 0 && (
                    <div style={{ marginTop: 16, textAlign: 'center' }}>
                        <div style={{ display: 'flex', alignItems: 'center', margin: '24px 0' }}>
                            <div style={{ flex: 1, height: '1px', backgroundColor: '#e8e8e8' }} />
                            <span style={{ margin: '0 16px', color: '#999', fontSize: '14px' }}>其他登录方式</span>
                            <div style={{ flex: 1, height: '1px', backgroundColor: '#e8e8e8' }} />
                        </div>
                        <Space size={[16, 24]} wrap style={{ display: 'flex', justifyContent: 'center', gap: '24px' }}>
                            {ssoConfigs.map(config => (
                                <Button
                                    key={config.name}
                                    style={{ backgroundColor: getRandomColor(), border: 'none' }}
                                    loading={loadingSSO[config.name]}
                                    onClick={() => {
                                        setLoadingSSO(prev => ({ ...prev, [config.name]: true }));
                                        window.location.href = `/auth/${config.type}/${config.name}/sso`;
                                    }}
                                >
                                    {config.name}
                                </Button>
                            ))}
                        </Space>
                    </div>
                )}
            </Form>
        </div>
    </section>
}

export default Login

const getRandomColor = () => {
    const hue = Math.floor(Math.random() * 360);
    return `hsl(${hue}, 70%, 65%)`;
};
