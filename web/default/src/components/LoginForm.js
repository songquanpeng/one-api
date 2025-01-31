import React, { useContext, useEffect, useState } from 'react';
import { Button, Divider, Form, Grid, Header, Image, Message, Modal, Segment } from 'semantic-ui-react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { UserContext } from '../context/User';
import { API, getLogo, showError, showSuccess, showWarning } from '../helpers';
import { onGitHubOAuthClicked, onLarkOAuthClicked } from './utils';
import { BASE_URL } from '../config';
import larkIcon from '../images/lark.svg';

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
    }
  }, []);

  const [showWeChatLoginModal, setShowWeChatLoginModal] = useState(false);

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
      const res = await API.post(`/api/user/login`, {
        username,
        password
      });
      const { success, message, data } = res.data;
      if (success) {
        userDispatch({ type: 'login', payload: data });
        localStorage.setItem('user', JSON.stringify(data));
        if (username === 'root' && password === '123456') {
          navigate('/user/edit');
          showSuccess('登录成功！');
          showWarning('请立刻修改默认密码！');
        } else {
          navigate('/token');
          showSuccess('登录成功！');
        }
      } else {
        showError(message);
      }
    }
  }

  return (
    <Grid textAlign='center' style={{ marginTop: '48px' }}>
      <Grid.Column style={{ maxWidth: 450 }}>
        <Header as='h2' color='' textAlign='center'>
          <Image src={ BASE_URL + logo} /> 用户登录
        </Header>
        <Form size='large'>
          <Segment>
            <Form.Input
              fluid
              icon='user'
              iconPosition='left'
              placeholder='用户名 / 邮箱地址'
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
        {status.github_oauth || status.wechat_login || status.lark_client_id ? (
          <>
            <Divider horizontal>Or</Divider>
            <div style={{ display: "flex", justifyContent: "center" }}>
              {status.github_oauth ? (
                <Button
                  circular
                  color='black'
                  icon='github'
                  onClick={() => onGitHubOAuthClicked(status.github_client_id)}
                />
              ) : (
                <></>
              )}
              {status.wechat_login ? (
                <Button
                  circular
                  color='green'
                  icon='wechat'
                  onClick={onWeChatLoginClicked}
                />
              ) : (
                <></>
              )}
              {status.lark_client_id ? (
                <div style={{
                    background: "radial-gradient(circle, #FFFFFF, #FFFFFF, #00D6B9, #2F73FF, #0a3A9C)",
                    width: "36px",
                    height: "36px",
                    borderRadius: "10em",
                    display: "flex",
                    cursor: "pointer"
                  }}
                  onClick={() => onLarkOAuthClicked(status.lark_client_id)}
                >
                  <Image
                    src={larkIcon}
                    avatar
                    style={{ width: "16px", height: "16px", cursor: "pointer", margin: "auto" }}
                    onClick={() => onLarkOAuthClicked(status.lark_client_id)}
                  />
                </div>
              ) : (
                <></>
              )}
            </div>
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
