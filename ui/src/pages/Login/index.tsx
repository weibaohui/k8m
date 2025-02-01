import {Form, Input, Button, Checkbox, Message} from '@arco-design/web-react'
import {useNavigate} from 'react-router-dom'
import {uuid} from '@/utils/utils'
import {
    IconUser,
    IconLock
} from '@arco-design/web-react/icon'
import styles from './index.module.scss'
import {useCallback} from 'react'

const FormItem = Form.Item

const Login = () => {
    const navigate = useNavigate()
    const [form] = Form.useForm();
    const onSubmit = useCallback(() => {
        form.validate().then(res => {
            if (res.username === 'admin' && res.password === '123456') {
                Message.success('登录成功')
                localStorage.setItem('token', uuid())
                navigate('/')
            }
        })
    }, [navigate, form])
    return <section className={styles.login}>
        <div className={styles.content}>
            <Form
                form={form}
                initialValues={{
                    username: 'admin',
                    password: '123456'
                }}
                className={styles.form} autoComplete='off'>
                <div>
                    <h2 style={{color: '#666', fontSize: '24px', marginBottom: 20}}>欢迎登录</h2>
                </div>
                <FormItem field={'username'} rules={[{required: true}]}>
                    <Input placeholder='请输入用户名' prefix={<IconUser/>}/>
                </FormItem>
                <FormItem field={'password'} rules={[{required: true}]}>
                    <Input.Password
                        prefix={<IconLock/>}
                        defaultVisibility={false}
                        placeholder='请输入密码'
                    />
                </FormItem>
                <FormItem>
                    <Checkbox>记住密码</Checkbox>
                </FormItem>
                <FormItem>
                    <Button type='primary' long onClick={onSubmit}>登 录</Button>
                </FormItem>
            </Form>
        </div>
    </section>
}

export default Login