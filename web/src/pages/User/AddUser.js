import React, { useState } from 'react';
import { Button, Form, Header, Segment } from 'semantic-ui-react';
import { API, showError, showSuccess } from '../../helpers';

const AddUser = () => {
  const originInputs = {
    username: '',
    display_name: '',
    password: '',
  };
  const [inputs, setInputs] = useState(originInputs);
  const { username, display_name, password } = inputs;

  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const submit = async () => {
    if (inputs.username === '' || inputs.password === '') return;
    const res = await API.post(`/api/user/`, inputs);
    const { success, message } = res.data;
    if (success) {
      showSuccess('User account created successfully!');
      setInputs(originInputs);
    } else {
      showError(message);
    }
  };

  return (
    <>
      <Segment>
        <Header as="h3">Create new user account</Header>
        <Form autoComplete="off">
          <Form.Field>
            <Form.Input
              label="Username"
              name="username"
              placeholder={'Please enter a username'}
              onChange={handleInputChange}
              value={username}
              autoComplete="off"
              required
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label="Display name"
              name="display_name"
              placeholder={'Please enter a display name'}
              onChange={handleInputChange}
              value={display_name}
              autoComplete="off"
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label="Password"
              name="password"
              type={'password'}
              placeholder={'Please enter a password!'}
              onChange={handleInputChange}
              value={password}
              autoComplete="off"
              required
            />
          </Form.Field>
          <Button type={'submit'} onClick={submit}>
            Submit
          </Button>
        </Form>
      </Segment>
    </>
  );
};

export default AddUser;
