import React, { useEffect, useState } from 'react';
import { Button, Form, Header, Message, Segment } from 'semantic-ui-react';
import { useParams } from 'react-router-dom';
import { API, showError, showSuccess } from '../../helpers';
import { CHANNEL_OPTIONS } from '../../constants';

const EditChannel = () => {
  const params = useParams();
  const channelId = params.id;
  const isEdit = channelId !== undefined;
  const [loading, setLoading] = useState(isEdit);
  const originInputs = {
    name: '',
    type: 1,
    key: '',
    base_url: ''
  };
  const [inputs, setInputs] = useState(originInputs);
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
    if (isEdit) {
      loadChannel().then();
    }
  }, []);

  const submit = async () => {
    if (!isEdit && (inputs.name === '' || inputs.key === '')) return;
    let localInputs = inputs;
    if (localInputs.base_url.endsWith('/')) {
      localInputs.base_url = localInputs.base_url.slice(0, localInputs.base_url.length - 1);
    }
    let res;
    if (isEdit) {
      res = await API.put(`/api/channel/`, { ...localInputs, id: parseInt(channelId) });
    } else {
      res = await API.post(`/api/channel/`, localInputs);
    }
    const { success, message } = res.data;
    if (success) {
      if (isEdit) {
        showSuccess('渠道更新成功！');
      } else {
        showSuccess('渠道创建成功！');
        setInputs(originInputs);
      }
    } else {
      showError(message);
    }
  };

  return (
    <>
      <Segment loading={loading}>
        <Header as='h3'>{isEdit ? '更新渠道信息' : '创建新的渠道'}</Header>
        <Form autoComplete='new-password'>
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
            inputs.type === 3 && (
              <>
                <Message>
                  注意，创建资源时，部署名称必须模型名称保持一致，因为 One API 会把请求体中的 model 参数替换为你的部署名称（模型名称中的点会被剔除）。
                </Message>
                <Form.Field>
                  <Form.Input
                    label='AZURE_OPENAI_ENDPOINT'
                    name='base_url'
                    placeholder={'请输入 AZURE_OPENAI_ENDPOINT，例如：https://docs-test-001.openai.azure.com'}
                    onChange={handleInputChange}
                    value={inputs.base_url}
                    autoComplete='new-password'
                  />
                </Form.Field>
              </>
            )
          }
          {
            inputs.type === 8 && (
              <Form.Field>
                <Form.Input
                  label='Base URL'
                  name='base_url'
                  placeholder={'请输入自定义渠道的 Base URL，例如：https://openai.justsong.cn'}
                  onChange={handleInputChange}
                  value={inputs.base_url}
                  autoComplete='new-password'
                />
              </Form.Field>
            )
          }
          <Form.Field>
            <Form.Input
              label='名称'
              name='name'
              placeholder={'请输入名称'}
              onChange={handleInputChange}
              value={inputs.name}
              autoComplete='new-password'
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label='密钥'
              name='key'
              placeholder={'请输入密钥'}
              onChange={handleInputChange}
              value={inputs.key}
              // type='password'
              autoComplete='new-password'
            />
          </Form.Field>
          <Button onClick={submit}>提交</Button>
        </Form>
      </Segment>
    </>
  );
};

export default EditChannel;
