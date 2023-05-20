import React, { useEffect, useState } from 'react';
import {
  Button,
  Form,
  Grid,
  Header,
  Image,
  Message,
  Segment,
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
      showInfo('Password must be at least 8 characters in length！');
      return;
    }
    if (password !== password2) {
      showInfo('The two entered passwords do not match.');
      return;
    }
    if (username && password) {
      if (turnstileEnabled && turnstileToken === '') {
        showInfo('Please wait for a few seconds and try again, Turnstile is checking the user environment！');
        return;
      }
      setLoading(true);
      const res = await API.post(
        `/api/user/register?turnstile=${turnstileToken}`,
        inputs
      );
      const { success, message } = res.data;
      if (success) {
        navigate('/login');
        showSuccess('Registration successful！');
      } else {
        showError(message);
      }
      setLoading(false);
    }
  }

  const sendVerificationCode = async () => {
    if (inputs.email === '') return;
    if (turnstileEnabled && turnstileToken === '') {
      showInfo('Please wait a few seconds and try again, Turnstile is checking the user environment！');
      return;
    }
    setLoading(true);
    const res = await API.get(
      `/api/verification?email=${inputs.email}&turnstile=${turnstileToken}`
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess('Verification code sent successfully, please check your email！');
    } else {
      showError(message);
    }
    setLoading(false);
  };

  return (
    <Grid textAlign='center' style={{ marginTop: '48px' }}>
      <Grid.Column style={{ maxWidth: 450 }}>
        <Header as='h2' color='' textAlign='center'>
          <Image src={logo} /> New User Registration
        </Header>
        <Form size='large'>
          <Segment>
            <Form.Input
              fluid
              icon='user'
              iconPosition='left'
              placeholder='Enter your username, up to 12 characters.'
              onChange={handleChange}
              name='username'
            />
            <Form.Input
              fluid
              icon='lock'
              iconPosition='left'
              placeholder='Enter your password with a minimum of 8 characters and a maximum of 20 characters.'
              onChange={handleChange}
              name='password'
              type='password'
            />
            <Form.Input
              fluid
              icon='lock'
              iconPosition='left'
              placeholder='Enter your password, with a minimum of 8 characters and a maximum of 20 characters.'
              onChange={handleChange}
              name='password2'
              type='password'
            />
            {showEmailVerification ? (
              <>
                <Form.Input
                  fluid
                  icon='mail'
                  iconPosition='left'
                  placeholder='Enter your email address'
                  onChange={handleChange}
                  name='email'
                  type='email'
                  action={
                    <Button onClick={sendVerificationCode} disabled={loading}>
                      Get Verification Code
                    </Button>
                  }
                />
                <Form.Input
                  fluid
                  icon='lock'
                  iconPosition='left'
                  placeholder='Enter verification code'
                  onChange={handleChange}
                  name='verification_code'
                />
              </>
            ) : (
              <></>
            )}
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
            <Button
              color=''
              fluid
              size='large'
              onClick={handleSubmit}
              loading={loading}
            >
              Registration
            </Button>
          </Segment>
        </Form>
        <Message>
          Already have an account?
          <Link to='/login' className='btn btn-link'>
            Click to login
          </Link>
        </Message>
      </Grid.Column>
    </Grid>
  );
};

export default RegisterForm;
