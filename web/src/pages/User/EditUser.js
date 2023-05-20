import React, { useEffect, useState } from 'react';
import { Button, Form, Header, Segment } from 'semantic-ui-react';
import { useParams } from 'react-router-dom';
import { API, showError, showSuccess } from '../../helpers';

const EditUser = () => {
  const params = useParams();
  const userId = params.id;
  const [loading, setLoading] = useState(true);
  const [inputs, setInputs] = useState({
    username: '',
    display_name: '',
    password: '',
    github_id: '',
    wechat_id: '',
    email: '',
  });
  const { username, display_name, password, github_id, wechat_id, email } =
    inputs;
  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const loadUser = async () => {
    let res = undefined;
    if (userId) {
      res = await API.get(`/api/user/${userId}`);
    } else {
      res = await API.get(`/api/user/self`);
    }
    const { success, message, data } = res.data;
    if (success) {
      data.password = '';
      setInputs(data);
    } else {
      showError(message);
    }
    setLoading(false);
  };
  useEffect(() => {
    loadUser().then();
  }, []);

  const submit = async () => {
    let res = undefined;
    if (userId) {
      res = await API.put(`/api/user/`, { ...inputs, id: parseInt(userId) });
    } else {
      res = await API.put(`/api/user/self`, inputs);
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess('User information updated successfully!');
    } else {
      showError(message);
    }
  };

  return (
    <>
      <Segment loading={loading}>
        <Header as='h3'>Update User Information</Header>
        <Form autoComplete='new-password'>
          <Form.Field>
            <Form.Input
              label='Username'
              name='username'
              placeholder={'Please enter a new username'}
              onChange={handleInputChange}
              value={username}
              autoComplete='new-password'
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label='Password'
              name='password'
              type={'password'}
              placeholder={'Please enter a new password'}
              onChange={handleInputChange}
              value={password}
              autoComplete='new-password'
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label='Display name'
              name='display_name'
              placeholder={'Please enter a new display name'}
              onChange={handleInputChange}
              value={display_name}
              autoComplete='new-password'
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label='Connected Github Account'
              name='github_id'
              value={github_id}
              autoComplete='new-password'
              placeholder='This setting is read-only. To change the connected account, please use the button on the personal settings page.'
              readOnly
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label='Connected Wechat Account'
              name='wechat_id'
              value={wechat_id}
              autoComplete='new-password'
              placeholder='This setting is read-only. To change the connected account, please use the button on the personal settings page.'
              readOnly
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label='Connected Email Account'
              name='email'
              value={email}
              autoComplete='new-password'
              placeholder='This setting is read-only. To change the connected account, please use the button on the personal settings page.'
              readOnly
            />
          </Form.Field>
          <Button onClick={submit}>Submit</Button>
        </Form>
      </Segment>
    </>
  );
};

export default EditUser;
