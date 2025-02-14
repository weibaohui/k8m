import { Form, Input, Button, Checkbox, message } from 'antd'
import { useNavigate } from 'react-router-dom'
import {
    UserOutlined,
    LockOutlined
} from '@ant-design/icons'
import styles from './index.module.scss'
import { useCallback, useEffect } from 'react'

// @ts-ignore
import CryptoJS from "crypto-js";

const FormItem = Form.Item

const secretKey = 'secret-key-16-ok';

// 加密函数
function encrypt(message: string) {
    // key 和 iv 使用同一个值
    const sKey = CryptoJS.enc.Utf8.parse(secretKey);
    const encrypted = CryptoJS.AES.encrypt(message, sKey, {
        iv: sKey,
        mode: CryptoJS.mode.CBC, // CBC算法
        padding: CryptoJS.pad.Pkcs7, //使用pkcs7 进行padding 后端需要注意
    });

    return encrypted.toString();
}

// 解密函数
function decrypt(base64CipherText: string) {
    // key 和 iv 使用同一个值
    const sKey = CryptoJS.enc.Utf8.parse(secretKey);
    const decrypted = CryptoJS.AES.decrypt(base64CipherText, sKey, {
        iv: sKey,
        mode: CryptoJS.mode.CBC, // CBC算法
        padding: CryptoJS.pad.Pkcs7, //使用pkcs7 进行padding 后端需要注意
    });

    return decrypted.toString(CryptoJS.enc.Utf8);
}

const Login = () => {
    const navigate = useNavigate()
    const [form] = Form.useForm();

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
    }, [navigate, form]);

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
                <FormItem name='remember' valuePropName='checked'>
                    <Checkbox>记住</Checkbox>
                </FormItem>
                <FormItem>
                    <Button type='primary' block onClick={onSubmit}>登 录</Button>
                </FormItem>
            </Form>
        </div>
    </section>
}

export default Login
