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
      showSuccess('用户账户创建成功！');
      setInputs(originInputs);
    } else {
      showError(message);
    }
  };

  return (
    <>
      <Segment>
        <Header as="h3">创建新用户账户</Header>
        <Form autoComplete="off">
          <Form.Field>
            <Form.Input
              label="用户名"
              name="username"
              placeholder={'请输入用户名'}
              onChange={handleInputChange}
              value={username}
              autoComplete="off"
              required
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label="显示名称"
              name="display_name"
              placeholder={'请输入显示名称'}
              onChange={handleInputChange}
              value={display_name}
              autoComplete="off"
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label="密码"
              name="password"
              type={'password'}
              placeholder={'请输入密码'}
              onChange={handleInputChange}
              value={password}
              autoComplete="off"
              required
            />
          </Form.Field>
          <Button positive type={'submit'} onClick={submit}>
            提交
          </Button>
        </Form>
      </Segment>
    </>
  );
};

export default AddUser;
