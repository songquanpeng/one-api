import React, { useState } from 'react';
import { Button, Form, Header, Segment } from 'semantic-ui-react';
import { API, showError, showSuccess } from '../../helpers';
import { CHANNEL_OPTIONS } from '../../constants';

const AddChannel = () => {
  const originInputs = {
    name: '',
    type: 1,
    key: ''
  };
  const [inputs, setInputs] = useState(originInputs);
  const { name, type, key } = inputs;

  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const submit = async () => {
    if (inputs.name === '' || inputs.key === '') return;
    const res = await API.post(`/api/channel/`, inputs);
    const { success, message } = res.data;
    if (success) {
      showSuccess('渠道创建成功！');
      setInputs(originInputs);
    } else {
      showError(message);
    }
  };

  return (
    <>
      <Segment>
        <Header as='h3'>创建新的渠道</Header>
        <Form autoComplete='off'>
          <Form.Field>
            <Form.Select
              label='类型'
              name='type'
              options={CHANNEL_OPTIONS}
              value={inputs.type}
              onChange={handleInputChange}
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label='名称'
              name='name'
              placeholder={'请输入名称'}
              onChange={handleInputChange}
              value={name}
              autoComplete='off'
              required
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label='密钥'
              name='key'
              placeholder={'请输入密钥'}
              onChange={handleInputChange}
              value={key}
              // type='password'
              autoComplete='off'
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

export default AddChannel;
