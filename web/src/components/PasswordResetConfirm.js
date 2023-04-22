import React, { useEffect, useState } from 'react';
import { Button, Form, Grid, Header, Image, Segment } from 'semantic-ui-react';
import { API, copy, showError, showSuccess } from '../helpers';
import { useSearchParams } from 'react-router-dom';

const PasswordResetConfirm = () => {
  const [inputs, setInputs] = useState({
    email: '',
    token: '',
  });
  const { email, token } = inputs;

  const [loading, setLoading] = useState(false);

  const [searchParams, setSearchParams] = useSearchParams();
  useEffect(() => {
    let token = searchParams.get('token');
    let email = searchParams.get('email');
    setInputs({
      token,
      email,
    });
  }, []);

  async function handleSubmit(e) {
    if (!email) return;
    setLoading(true);
    const res = await API.post(`/api/user/reset`, {
      email,
      token,
    });
    const { success, message } = res.data;
    if (success) {
      let password = res.data.data;
      await copy(password);
      showSuccess(`密码已重置并已复制到剪贴板：${password}`);
    } else {
      showError(message);
    }
    setLoading(false);
  }

  return (
    <Grid textAlign='center' style={{ marginTop: '48px' }}>
      <Grid.Column style={{ maxWidth: 450 }}>
        <Header as='h2' color='' textAlign='center'>
          <Image src='/logo.png' /> 密码重置确认
        </Header>
        <Form size='large'>
          <Segment>
            <Form.Input
              fluid
              icon='mail'
              iconPosition='left'
              placeholder='邮箱地址'
              name='email'
              value={email}
              readOnly
            />
            <Button
              color=''
              fluid
              size='large'
              onClick={handleSubmit}
              loading={loading}
            >
              提交
            </Button>
          </Segment>
        </Form>
      </Grid.Column>
    </Grid>
  );
};

export default PasswordResetConfirm;
