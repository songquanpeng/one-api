import React, { useContext, useEffect, useState } from 'react';
import {
  Button,
  Divider,
  Form,
  Grid,
  Header,
  Image,
  Message,
  Modal,
  Segment,
} from 'semantic-ui-react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { UserContext } from '../context/User';
import { API, getLogo, showError, showSuccess } from '../helpers';

const LoginForm = () => {
  const [inputs, setInputs] = useState({
    username: '',
    password: '',
    wechat_verification_code: '',
  });
  const [searchParams, setSearchParams] = useSearchParams();
  const [submitted, setSubmitted] = useState(false);
  const { username, password } = inputs;
  const [userState, userDispatch] = useContext(UserContext);
  let navigate = useNavigate();

  const [status, setStatus] = useState({});
  const logo = getLogo();

  useEffect(() => {
    if (searchParams.get("expired")) {
      showError('Not logged in or session has expired, please log in again！');
    }
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      setStatus(status);
    }
  }, []);

  const [showWeChatLoginModal, setShowWeChatLoginModal] = useState(false);

  const onGitHubOAuthClicked = () => {
    window.open(
      `https://github.com/login/oauth/authorize?client_id=${status.github_client_id}&scope=user:email`
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
      showSuccess('Login successful！');
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
      const res = await API.post('/api/user/login', {
        username,
        password,
      });
      const { success, message, data } = res.data;
      if (success) {
        userDispatch({ type: 'login', payload: data });
        localStorage.setItem('user', JSON.stringify(data));
        navigate('/');
        showSuccess('Login successful！');
      } else {
        showError(message);
      }
    }
  }

  return (
    <Grid textAlign="center" style={{ marginTop: '48px' }}>
      <Grid.Column style={{ maxWidth: 450 }}>
        <Header as="h2" color="" textAlign="center">
          <Image src={logo} /> Login
        </Header>
        <Form size="large">
          <Segment>
            <Form.Input
              fluid
              icon="user"
              iconPosition="left"
              placeholder="Username"
              name="username"
              value={username}
              onChange={handleChange}
            />
            <Form.Input
              fluid
              icon="lock"
              iconPosition="left"
              placeholder="Password"
              name="password"
              type="password"
              value={password}
              onChange={handleChange}
            />
            <Button color="" fluid size="large" onClick={handleSubmit}>
              Login
            </Button>
          </Segment>
        </Form>
        <Message>
          Forgot Password 
          <Link to="/reset" className="btn btn-link">
            Reset Password
          </Link>
          ； No account？
          <Link to="/register" className="btn btn-link">
            Click to Register
          </Link>
        </Message>
        {status.github_oauth || status.wechat_login ? (
          <>
            <Divider horizontal>Or</Divider>
            {status.github_oauth ? (
              <Button
                circular
                color="black"
                icon="github"
                onClick={onGitHubOAuthClicked}
              />
            ) : (
              <></>
            )}
            {status.wechat_login ? (
              <Button
                circular
                color="green"
                icon="wechat"
                onClick={onWeChatLoginClicked}
              />
            ) : (
              <></>
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
                  Scan the QR code to follow our official WeChat account and enter the "verification code" to obtain a verification code (valid for three minutes).
                </p>
              </div>
              <Form size="large">
                <Form.Input
                  fluid
                  placeholder="Verification code"
                  name="wechat_verification_code"
                  value={inputs.wechat_verification_code}
                  onChange={handleChange}
                />
                <Button
                  color=""
                  fluid
                  size="large"
                  onClick={onSubmitWeChatVerificationCode}
                >
                  Login
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
