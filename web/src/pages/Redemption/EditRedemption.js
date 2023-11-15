import React, { useEffect, useState } from 'react';
import { Button, Form, Header, Segment } from 'semantic-ui-react';
import { useParams, useNavigate } from 'react-router-dom';
import { API, downloadTextAsFile, showError, showSuccess } from '../../helpers';
import { renderQuota, renderQuotaWithPrompt } from '../../helpers/render';

const EditRedemption = () => {
  const params = useParams();
  const navigate = useNavigate();
  const redemptionId = params.id;
  const isEdit = redemptionId !== undefined;
  const [loading, setLoading] = useState(isEdit);
  const originInputs = {
    name: '',
    quota: 100000,
    count: 1
  };
  const [inputs, setInputs] = useState(originInputs);
  const { name, quota, count } = inputs;

  const handleCancel = () => {
    navigate('/redemption');
  };
  
  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const loadRedemption = async () => {
    let res = await API.get(`/api/redemption/${redemptionId}`);
    const { success, message, data } = res.data;
    if (success) {
      setInputs(data);
    } else {
      showError(message);
    }
    setLoading(false);
  };
  useEffect(() => {
    if (isEdit) {
      loadRedemption().then();
    }
  }, []);

  const submit = async () => {
    if (!isEdit && inputs.name === '') return;
    let localInputs = inputs;
    localInputs.count = parseInt(localInputs.count);
    localInputs.quota = parseInt(localInputs.quota);
    let res;
    if (isEdit) {
      res = await API.put(`/api/redemption/`, { ...localInputs, id: parseInt(redemptionId) });
    } else {
      res = await API.post(`/api/redemption/`, {
        ...localInputs
      });
    }
    const { success, message, data } = res.data;
    if (success) {
      if (isEdit) {
        showSuccess('兑换码更新成功！');
      } else {
        showSuccess('兑换码创建成功！');
        setInputs(originInputs);
      }
    } else {
      showError(message);
    }
    if (!isEdit && data) {
      let text = "";
      for (let i = 0; i < data.length; i++) {
        text += data[i] + "\n";
      }
      downloadTextAsFile(text, `${inputs.name}.txt`);
    }
  };

  return (
    <>
      <Segment loading={loading}>
        <Header as='h3'>{isEdit ? '更新兑换码信息' : '创建新的兑换码'}</Header>
        <Form autoComplete='new-password'>
          <Form.Field>
            <Form.Input
              label='名称'
              name='name'
              placeholder={'请输入名称'}
              onChange={handleInputChange}
              value={name}
              autoComplete='new-password'
              required={!isEdit}
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label={`额度${renderQuotaWithPrompt(quota)}`}
              name='quota'
              placeholder={'请输入单个兑换码中包含的额度'}
              onChange={handleInputChange}
              value={quota}
              autoComplete='new-password'
              type='number'
            />
          </Form.Field>
          {
            !isEdit && <>
              <Form.Field>
                <Form.Input
                  label='生成数量'
                  name='count'
                  placeholder={'请输入生成数量'}
                  onChange={handleInputChange}
                  value={count}
                  autoComplete='new-password'
                  type='number'
                />
              </Form.Field>
            </>
          }
          <Button positive onClick={submit}>提交</Button>
          <Button onClick={handleCancel}>取消</Button>
        </Form>
      </Segment>
    </>
  );
};

export default EditRedemption;
