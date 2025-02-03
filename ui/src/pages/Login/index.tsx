import { Form, Input, Button, Checkbox, Message } from '@arco-design/web-react'
import { useNavigate } from 'react-router-dom'
import {
    IconUser,
    IconLock
} from '@arco-design/web-react/icon'
import styles from './index.module.scss'
import { useCallback, useEffect } from 'react'

const FormItem = Form.Item

const Login = () => {
    const navigate = useNavigate()
    const [form] = Form.useForm();

    // useEffect 读取 remember 数据
    useEffect(() => {
        const savedData = localStorage.getItem('remember');
        if (savedData) {
            const parsedData = JSON.parse(savedData);
            form.setFieldsValue(parsedData);
            form.setFieldValue('remember', true);  // 确保 remember 的值为 boolean

        }
    }, [form]);
    const onSubmit = useCallback(() => {
        form.validate().then(async (values) => {
            try {
                const res = await fetch('/auth/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(values)
                });
                const data = await res.json();
                if (res.ok) {
                    Message.success('登录成功');
                    localStorage.setItem('token', data.token);
                    if (values.remember) {
                        localStorage.setItem('remember', JSON.stringify(values));
                    } else {
                        localStorage.removeItem('remember');
                    }
                    navigate('/');
                } else {
                    Message.error(data.message || '登录失败');
                }
            } catch (error) {
                Message.error('网络错误');
            }
        });
    }, [navigate, form]);
    return <section className={styles.login}>
        <div className={styles.content}>
            <Form
                form={form}
                // initialValues={{
                //     username: 'admin',
                //     password: '123456'
                // }}
                className={styles.form} autoComplete='off'>
                <div>
                    <h2 style={{ color: '#666', fontSize: '24px', marginBottom: 20 }}>欢迎登录</h2>
                </div>
                <FormItem field={'username'} rules={[{ required: true }]}>
                    <Input placeholder='请输入用户名' prefix={<IconUser />} />
                </FormItem>
                <FormItem field={'password'} rules={[{ required: true }]}>
                    <Input.Password
                        prefix={<IconLock />}
                        defaultVisibility={false}
                        placeholder='请输入密码'
                    />
                </FormItem>
                <FormItem field={'remember'} triggerPropName='checked' >
                    <Checkbox >记住</Checkbox>
                </FormItem>
                <FormItem>
                    <Button type='primary' long onClick={onSubmit}>登 录</Button>
                </FormItem>
            </Form>
        </div>
    </section>
}

export default Login
