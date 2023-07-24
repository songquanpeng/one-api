import React, { useContext, useEffect, useState } from 'react';
import { Button, Divider, Form, Grid, Header, Image, Message, Modal, Segment } from 'semantic-ui-react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { UserContext } from '../context/User';
import { API, getLogo, showError, showInfo, showSuccess } from '../helpers';
import Turnstile from 'react-turnstile';

const LoginForm = () => {
  const [inputs, setInputs] = useState({
    username: '',
    password: '',
    wechat_verification_code: ''
  });
  const [searchParams, setSearchParams] = useSearchParams();
  const [submitted, setSubmitted] = useState(false);
  const { username, password } = inputs;
  const [userState, userDispatch] = useContext(UserContext);
  const [turnstileEnabled, setTurnstileEnabled] = useState(false);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [turnstileToken, setTurnstileToken] = useState('');
  let navigate = useNavigate();
  const [status, setStatus] = useState({});
  const logo = getLogo();

  useEffect(() => {
    if (searchParams.get('expired')) {
      showError('未登录或登录已过期，请重新登录！');
    }
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      setStatus(status);

      if (status.turnstile_check) {
        setTurnstileEnabled(true);
        setTurnstileSiteKey(status.turnstile_site_key);
      }
    }
  }, []);

  const [showWeChatLoginModal, setShowWeChatLoginModal] = useState(false);

  const onGitHubOAuthClicked = () => {
    window.open(
      `https://github.com/login/oauth/authorize?client_id=${status.github_client_id}&scope=user:email`
    );
  };

  const onDiscordOAuthClicked = () => {
    window.open(
      `https://discord.com/oauth2/authorize?response_type=code&client_id=${status.discord_client_id}&redirect_uri=${window.location.origin}/oauth/discord&scope=identify`,
    );
  };

  const onWeChatLoginClicked = () => {
    setShowWeChatLoginModal(true);
  };

  const onSubmitWeChatVerificationCode = async () => {
    const res = await API.get(
      `/api/oauth/wechat?code=${inputs.wechat_verification_code}`
    );
    const { success, message, data } = res.data;
    if (success) {
      userDispatch({ type: 'login', payload: data });
      localStorage.setItem('user', JSON.stringify(data));
      navigate('/');
      showSuccess('登录成功！');
      setShowWeChatLoginModal(false);
    } else {
      showError(message);
    }
  };

  function handleChange(e) {
    const { name, value } = e.target;
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  }

  async function handleSubmit(e) {
    setSubmitted(true);
    if (username && password) {
      if (turnstileEnabled && turnstileToken === '') {
        showInfo('请稍后几秒重试，Turnstile 正在检查用户环境！');
        return;
      }

      const res = await API.post(`/api/user/login?turnstile=${turnstileToken}`, {
        username,
        password
      });
      const { success, message, data } = res.data;
      if (success) {
        userDispatch({ type: 'login', payload: data });
        localStorage.setItem('user', JSON.stringify(data));
        navigate('/');
        showSuccess('登录成功！');
      } else {
        showError(message);
      }
    }
  }

  return (
    <Grid textAlign='center' style={{ marginTop: '48px' }}>
      <Grid.Column style={{ maxWidth: 450 }}>
        <Header as='h2' color='' textAlign='center'>
          <Image src={logo} /> 用户登录
        </Header>
        <Form size='large'>
          <Segment>
            <Form.Input
              fluid
              icon='user'
              iconPosition='left'
              placeholder='用户名'
              name='username'
              value={username}
              onChange={handleChange}
            />
            <Form.Input
              fluid
              icon='lock'
              iconPosition='left'
              placeholder='密码'
              name='password'
              type='password'
              value={password}
              onChange={handleChange}
            />
            {turnstileEnabled ? (
              <Turnstile
                sitekey={turnstileSiteKey}
                onVerify={(token) => {
                  setTurnstileToken(token);
                }}
              />
            ) : (
              <></>
            )}
            <Button color='green' fluid size='large' onClick={handleSubmit}>
              登录
            </Button>
          </Segment>
        </Form>
        <Message>
          忘记密码？
          <Link to='/reset' className='btn btn-link'>
            点击重置
          </Link>
          ； 没有账户？
          <Link to='/register' className='btn btn-link'>
            点击注册
          </Link>
        </Message>
        {status.github_oauth || status.wechat_login || status.discord_oauth ? (
          <>
            <Divider horizontal>Or</Divider>
            {status.discord_oauth && (
              <Button
                circular
                color='blue'
                icon='discord'
                onClick={onDiscordOAuthClicked}
              />
            )}
            {status.github_oauth && (
              <Button
                circular
                color='black'
                icon='github'
                onClick={onGitHubOAuthClicked}
              />
            )}
            {status.wechat_login && (
              <Button
                circular
                color='green'
                icon='wechat'
                onClick={onWeChatLoginClicked}
              />
            )}
          </>
        ) : (
          <></>
        )}
        <Modal
          onClose={() => setShowWeChatLoginModal(false)}
          onOpen={() => setShowWeChatLoginModal(true)}
          open={showWeChatLoginModal}
          size={'mini'}
        >
          <Modal.Content>
            <Modal.Description>
              <Image src={status.wechat_qrcode} fluid />
              <div style={{ textAlign: 'center' }}>
                <p>
                  微信扫码关注公众号，输入「验证码」获取验证码（三分钟内有效）
                </p>
              </div>
              <Form size='large'>
                <Form.Input
                  fluid
                  placeholder='验证码'
                  name='wechat_verification_code'
                  value={inputs.wechat_verification_code}
                  onChange={handleChange}
                />
                <Button
                  color=''
                  fluid
                  size='large'
                  onClick={onSubmitWeChatVerificationCode}
                >
                  登录
                </Button>
              </Form>
            </Modal.Description>
          </Modal.Content>
        </Modal>
      </Grid.Column>
    </Grid>
  );
};

export default LoginForm;
