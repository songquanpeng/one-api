import React, { useEffect, useState } from 'react';
import { Button, Form, Header, Segment } from 'semantic-ui-react';
import { useParams } from 'react-router-dom';
import { API, showError, showSuccess } from '../../helpers';

const EditToken = () => {
  const params = useParams();
  const tokenId = params.id;
  const [loading, setLoading] = useState(true);
  const [inputs, setInputs] = useState({
    name: ''
  });
  const { name } = inputs;
  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const loadToken = async () => {
    let res = await API.get(`/api/token/${tokenId}`);
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
    loadToken().then();
  }, []);

  const submit = async () => {
    let res = await API.put(`/api/token/`, { ...inputs, id: parseInt(tokenId) });
    const { success, message } = res.data;
    if (success) {
      showSuccess('令牌更新成功！');
    } else {
      showError(message);
    }
  };

  return (
    <>
      <Segment loading={loading}>
        <Header as='h3'>更新令牌信息</Header>
        <Form autoComplete='off'>
          <Form.Field>
            <Form.Input
              label='名称'
              name='name'
              placeholder={'请输入新的名称'}
              onChange={handleInputChange}
              value={name}
              autoComplete='off'
            />
          </Form.Field>
          <Button onClick={submit}>提交</Button>
        </Form>
      </Segment>
    </>
  );
};

export default EditToken;
