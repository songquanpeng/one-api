import React, { useEffect, useState } from 'react';
import { Button, Form, Header, Segment } from 'semantic-ui-react';
import { useParams } from 'react-router-dom';
import { API, showError, showSuccess } from '../../helpers';
import { CHANNEL_OPTIONS } from '../../constants';

const EditChannel = () => {
  const params = useParams();
  const channelId = params.id;
  const [loading, setLoading] = useState(true);
  const [inputs, setInputs] = useState({
    name: '',
    key: '',
    type: 1,
    base_url: '',
  });
  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const loadChannel = async () => {
    let res = await API.get(`/api/channel/${channelId}`);
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
    loadChannel().then();
  }, []);

  const submit = async () => {
    if (inputs.base_url.endsWith('/')) {
      inputs.base_url = inputs.base_url.slice(0, inputs.base_url.length - 1);
    }
    let res = await API.put(`/api/channel/`, { ...inputs, id: parseInt(channelId) });
    const { success, message } = res.data;
    if (success) {
      showSuccess('渠道更新成功！');
    } else {
      showError(message);
    }
  };

  return (
    <>
      <Segment loading={loading}>
        <Header as='h3'>更新渠道信息</Header>
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
          {
            inputs.type === 8 && (
              <Form.Field>
                <Form.Input
                  label='Base URL'
                  name='base_url'
                  placeholder={'请输入新的自定义渠道的 Base URL'}
                  onChange={handleInputChange}
                  value={inputs.base_url}
                  autoComplete='off'
                />
              </Form.Field>
            )
          }
          <Form.Field>
            <Form.Input
              label='名称'
              name='name'
              placeholder={'请输入新的名称'}
              onChange={handleInputChange}
              value={inputs.name}
              autoComplete='off'
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label='密钥'
              name='key'
              placeholder={'请输入新的密钥'}
              onChange={handleInputChange}
              value={inputs.key}
              // type='password'
              autoComplete='off'
            />
          </Form.Field>
          <Button onClick={submit}>提交</Button>
        </Form>
      </Segment>
    </>
  );
};

export default EditChannel;
