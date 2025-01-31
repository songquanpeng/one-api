import React, { useEffect, useState } from 'react';
import {
  Button,
  Form,
  Grid,
  Header,
  Image,
  Message,
  Segment,
  Card,
  Divider,
} from 'semantic-ui-react';
import { Link, useNavigate } from 'react-router-dom';
import { API, getLogo, showError, showInfo, showSuccess } from '../helpers';
import Turnstile from 'react-turnstile';

const RegisterForm = () => {
  const [inputs, setInputs] = useState({
    username: '',
    password: '',
    password2: '',
    email: '',
    verification_code: '',
  });
  const { username, password, password2 } = inputs;
  const [showEmailVerification, setShowEmailVerification] = useState(false);
  const [turnstileEnabled, setTurnstileEnabled] = useState(false);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [turnstileToken, setTurnstileToken] = useState('');
  const [loading, setLoading] = useState(false);
  const logo = getLogo();
  let affCode = new URLSearchParams(window.location.search).get('aff');
  if (affCode) {
    localStorage.setItem('aff', affCode);
  }

  useEffect(() => {
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      setShowEmailVerification(status.email_verification);
      if (status.turnstile_check) {
        setTurnstileEnabled(true);
        setTurnstileSiteKey(status.turnstile_site_key);
      }
    }
  });

  let navigate = useNavigate();

  function handleChange(e) {
    const { name, value } = e.target;
    console.log(name, value);
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  }

  async function handleSubmit(e) {
    if (password.length < 8) {
      showInfo('密码长度不得小于 8 位！');
      return;
    }
    if (password !== password2) {
      showInfo('两次输入的密码不一致');
      return;
    }
    if (username && password) {
      if (turnstileEnabled && turnstileToken === '') {
        showInfo('请稍后几秒重试，Turnstile 正在检查用户环境！');
        return;
      }
      setLoading(true);
      if (!affCode) {
        affCode = localStorage.getItem('aff');
      }
      inputs.aff_code = affCode;
      const res = await API.post(
        `/api/user/register?turnstile=${turnstileToken}`,
        inputs
      );
      const { success, message } = res.data;
      if (success) {
        navigate('/login');
        showSuccess('注册成功！');
      } else {
        showError(message);
      }
      setLoading(false);
    }
  }

  const sendVerificationCode = async () => {
    if (inputs.email === '') return;
    if (turnstileEnabled && turnstileToken === '') {
      showInfo('请稍后几秒重试，Turnstile 正在检查用户环境！');
      return;
    }
    setLoading(true);
    const res = await API.get(
      `/api/verification?email=${inputs.email}&turnstile=${turnstileToken}`
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess('验证码发送成功，请检查你的邮箱！');
    } else {
      showError(message);
    }
    setLoading(false);
  };

  return (
    <Grid textAlign='center' style={{ marginTop: '48px' }}>
      <Grid.Column style={{ maxWidth: 450 }}>
        <Card
          fluid
          className='chart-card'
          style={{ boxShadow: '0 1px 3px rgba(0,0,0,0.12)' }}
        >
          <Card.Content>
            <Card.Header>
              <Header
                as='h2'
                textAlign='center'
                style={{ marginBottom: '1.5em' }}
              >
                <Image src={logo} style={{ marginBottom: '10px' }} />
                <Header.Content>新用户注册</Header.Content>
              </Header>
            </Card.Header>
            <Form size='large'>
              <Form.Input
                fluid
                icon='user'
                iconPosition='left'
                placeholder='输入用户名，最长 12 位'
                onChange={handleChange}
                name='username'
                style={{ marginBottom: '1em' }}
              />
              <Form.Input
                fluid
                icon='lock'
                iconPosition='left'
                placeholder='输入密码，最短 8 位，最长 20 位'
                onChange={handleChange}
                name='password'
                type='password'
                style={{ marginBottom: '1em' }}
              />
              <Form.Input
                fluid
                icon='lock'
                iconPosition='left'
                placeholder='再次输入密码'
                onChange={handleChange}
                name='password2'
                type='password'
                style={{ marginBottom: '1em' }}
              />

              {showEmailVerification && (
                <>
                  <Form.Input
                    fluid
                    icon='mail'
                    iconPosition='left'
                    placeholder='输入邮箱地址'
                    onChange={handleChange}
                    name='email'
                    type='email'
                    action={
                      <Button
                        onClick={sendVerificationCode}
                        disabled={loading}
                        style={{ backgroundColor: '#2185d0', color: 'white' }}
                      >
                        获取验证码
                      </Button>
                    }
                    style={{ marginBottom: '1em' }}
                  />
                  <Form.Input
                    fluid
                    icon='lock'
                    iconPosition='left'
                    placeholder='输入验证码'
                    onChange={handleChange}
                    name='verification_code'
                    style={{ marginBottom: '1em' }}
                  />
                </>
              )}

              {turnstileEnabled && (
                <div
                  style={{
                    marginBottom: '1em',
                    display: 'flex',
                    justifyContent: 'center',
                  }}
                >
                  <Turnstile
                    sitekey={turnstileSiteKey}
                    onVerify={(token) => {
                      setTurnstileToken(token);
                    }}
                  />
                </div>
              )}

              <Button
                color='blue'
                fluid
                size='large'
                onClick={handleSubmit}
                loading={loading}
              >
                注册
              </Button>
            </Form>

            <Divider />
            <Message style={{ background: 'transparent', boxShadow: 'none' }}>
              <div
                style={{
                  textAlign: 'center',
                  fontSize: '0.9em',
                  color: '#666',
                }}
              >
                已有账户？
                <Link to='/login' style={{ color: '#2185d0' }}>
                  点击登录
                </Link>
              </div>
            </Message>
          </Card.Content>
        </Card>
      </Grid.Column>
    </Grid>
  );
};

export default RegisterForm;
