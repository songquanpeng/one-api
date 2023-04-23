import React, { useState } from 'react';
import { Button, Form, Header, Segment } from 'semantic-ui-react';
import { API, showError, showSuccess } from '../../helpers';

const AddToken = () => {
  const originInputs = {
    name: '',
  };
  const [inputs, setInputs] = useState(originInputs);
  const { name, display_name, password } = inputs;

  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const submit = async () => {
    if (inputs.name === '') return;
    const res = await API.post(`/api/token/`, inputs);
    const { success, message } = res.data;
    if (success) {
      showSuccess('令牌创建成功！');
      setInputs(originInputs);
    } else {
      showError(message);
    }
  };

  return (
    <>
      <Segment>
        <Header as="h3">创建新的令牌</Header>
        <Form autoComplete="off">
          <Form.Field>
            <Form.Input
              label="名称"
              name="name"
              placeholder={'请输入名称'}
              onChange={handleInputChange}
              value={name}
              autoComplete="off"
              required
            />
          </Form.Field>
          <Button type={'submit'} onClick={submit}>
            提交
          </Button>
        </Form>
      </Segment>
    </>
  );
};

export default AddToken;
