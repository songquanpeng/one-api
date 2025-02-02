import React, { useEffect, useState } from 'react';
import {
  Button,
  Form,
  Grid,
  Header,
  Image,
  Message,
  Card,
  Divider,
} from 'semantic-ui-react';
import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { API, getLogo, showError, showInfo, showSuccess } from '../helpers';
import Turnstile from 'react-turnstile';

const RegisterForm = () => {
  const { t } = useTranslation();
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
  const [disableButton, setDisableButton] = useState(false);
  const [countdown, setCountdown] = useState(30);
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

  useEffect(() => {
    let countdownInterval = null;
    if (disableButton && countdown > 0) {
      countdownInterval = setInterval(() => {
        setCountdown(countdown - 1);
      }, 1000);
    } else if (countdown === 0) {
      setDisableButton(false);
      setCountdown(30);
    }
    return () => clearInterval(countdownInterval);
  }, [disableButton, countdown]);

  let navigate = useNavigate();

  function handleChange(e) {
    const { name, value } = e.target;
    console.log(name, value);
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  }

  async function handleSubmit(e) {
    if (password.length < 8) {
      showInfo(t('messages.error.password_length'));
      return;
    }
    if (password !== password2) {
      showInfo(t('messages.error.password_mismatch'));
      return;
    }
    if (username && password) {
      if (turnstileEnabled && turnstileToken === '') {
        showInfo(t('messages.error.turnstile_wait'));
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
        showSuccess(t('messages.success.register'));
      } else {
        showError(message);
      }
      setLoading(false);
    }
  }

  const sendVerificationCode = async () => {
    if (inputs.email === '') return;
    if (turnstileEnabled && turnstileToken === '') {
      showInfo(t('messages.error.turnstile_wait'));
      return;
    }
    setDisableButton(true);
    setLoading(true);
    const res = await API.get(
      `/api/verification?email=${inputs.email}&turnstile=${turnstileToken}`
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('messages.success.verification_code'));
    } else {
      showError(message);
      setDisableButton(false);
      setCountdown(30);
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
                <Header.Content>{t('auth.register.title')}</Header.Content>
              </Header>
            </Card.Header>
            <Form size='large'>
              <Form.Input
                fluid
                icon='user'
                iconPosition='left'
                placeholder={t('auth.register.username')}
                onChange={handleChange}
                name='username'
                style={{ marginBottom: '1em' }}
              />
              <Form.Input
                fluid
                icon='lock'
                iconPosition='left'
                placeholder={t('auth.register.password')}
                onChange={handleChange}
                name='password'
                type='password'
                style={{ marginBottom: '1em' }}
              />
              <Form.Input
                fluid
                icon='lock'
                iconPosition='left'
                placeholder={t('auth.register.confirm_password')}
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
                    placeholder={t('auth.register.email')}
                    onChange={handleChange}
                    name='email'
                    type='email'
                    action={
                      <Button
                        onClick={sendVerificationCode}
                        disabled={loading}
                      >
                        {disableButton 
                          ? t('auth.register.get_code_retry', { countdown })
                          : t('auth.register.get_code')}
                      </Button>
                    }
                    style={{ marginBottom: '1em' }}
                  />
                  <Form.Input
                    fluid
                    icon='lock'
                    iconPosition='left'
                    placeholder={t('auth.register.verification_code')}
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
                fluid
                size='large'
                onClick={handleSubmit}
                style={{
                  background: '#2F73FF', // 使用更现代的蓝色
                  color: 'white',
                  marginBottom: '1.5em',
                }}
                loading={loading}
              >
                {t('auth.register.button')}
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
                {t('auth.register.has_account')}
                <Link to='/login' style={{ color: '#2185d0' }}>
                  {t('auth.register.login')}
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
