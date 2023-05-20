import React, { useEffect, useState } from 'react';
import { Button, Form, Header, Message, Segment } from 'semantic-ui-react';
import { useParams } from 'react-router-dom';
import { API, showError, showSuccess, timestamp2string } from '../../helpers';

const EditToken = () => {
  const params = useParams();
  const tokenId = params.id;
  const isEdit = tokenId !== undefined;
  const [loading, setLoading] = useState(isEdit);
  const originInputs = {
    name: '',
    remain_quota: 0,
    expired_time: -1,
    unlimited_quota: false
  };
  const [inputs, setInputs] = useState(originInputs);
  const { name, remain_quota, expired_time, unlimited_quota } = inputs;

  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const setExpiredTime = (month, day, hour, minute) => {
    let now = new Date();
    let timestamp = now.getTime() / 1000;
    let seconds = month * 30 * 24 * 60 * 60;
    seconds += day * 24 * 60 * 60;
    seconds += hour * 60 * 60;
    seconds += minute * 60;
    if (seconds !== 0) {
      timestamp += seconds;
      setInputs({ ...inputs, expired_time: timestamp2string(timestamp) });
    } else {
      setInputs({ ...inputs, expired_time: -1 });
    }
  };

  const setUnlimitedQuota = () => {
    setInputs({ ...inputs, unlimited_quota: !unlimited_quota });
  };

  const loadToken = async () => {
    let res = await API.get(`/api/token/${tokenId}`);
    const { success, message, data } = res.data;
    if (success) {
      if (data.expired_time !== -1) {
        data.expired_time = timestamp2string(data.expired_time);
      }
      setInputs(data);
    } else {
      showError(message);
    }
    setLoading(false);
  };
  useEffect(() => {
    if (isEdit) {
      loadToken().then();
    }
  }, []);

  const submit = async () => {
    if (!isEdit && inputs.name === '') return;
    let localInputs = inputs;
    localInputs.remain_quota = parseInt(localInputs.remain_quota);
    if (localInputs.expired_time !== -1) {
      let time = Date.parse(localInputs.expired_time);
      if (isNaN(time)) {
        showError('Expiration time format error!');
        return;
      }
      localInputs.expired_time = Math.ceil(time / 1000);
    }
    let res;
    if (isEdit) {
      res = await API.put(`/api/token/`, { ...localInputs, id: parseInt(tokenId) });
    } else {
      res = await API.post(`/api/token/`, localInputs);
    }
    const { success, message } = res.data;
    if (success) {
      if (isEdit) {
        showSuccess('Token updated successfully!');
      } else {
        showSuccess('Token created successfully!');
        setInputs(originInputs);
      }
    } else {
      showError(message);
    }
  };

  return (
    <>
      <Segment loading={loading}>
        <Header as='h3'>{isEdit ? 'Edit token' : 'Create a new token'}</Header>
        <Form autoComplete='new-password'>
          <Form.Field>
            <Form.Input
              label='Name'
              name='name'
              placeholder={'Please enter a name'}
              onChange={handleInputChange}
              value={name}
              autoComplete='new-password'
              required={!isEdit}
            />
          </Form.Field>
          <Message>Note that the amount of the token is only used to limit the maximum usage of the token itself, and the actual usage is limited by the remaining amount of the account.</Message>
          <Form.Field>
            <Form.Input
              label='Quota'
              name='remain_quota'
              placeholder={'Please enter an amount'}
              onChange={handleInputChange}
              value={remain_quota}
              autoComplete='new-password'
              type='number'
              disabled={unlimited_quota}
            />
          </Form.Field>
          <Button type={'button'} style={{ marginBottom: '14px' }} onClick={() => {
            setUnlimitedQuota();
          }}>{unlimited_quota ? 'Unlimited' : 'Set to unlimited'}</Button>
          <Form.Field>
            <Form.Input
              label='Expiration'
              name='expired_time'
              placeholder={'Please enter the expiration time in the format of yyyy-MM-dd HH:mm:ss, -1 means unlimited'}
              onChange={handleInputChange}
              value={expired_time}
              autoComplete='new-password'
              type='datetime-local'
            />
          </Form.Field>
          <Button type={'button'} onClick={() => {
            setExpiredTime(0, 0, 0, 0);
          }}>Never expires</Button>
          <Button type={'button'} onClick={() => {
            setExpiredTime(1, 0, 0, 0);
          }}>Expires in one month</Button>
          <Button type={'button'} onClick={() => {
            setExpiredTime(0, 1, 0, 0);
          }}>Expires in one day</Button>
          <Button type={'button'} onClick={() => {
            setExpiredTime(0, 0, 1, 0);
          }}>Expires in one hour</Button>
          <Button type={'button'} onClick={() => {
            setExpiredTime(0, 0, 0, 1);
          }}>Expires in one minute</Button>
          <br/>
          <Button onClick={submit}>Submit</Button>
        </Form>
      </Segment>
    </>
  );
};

export default EditToken;
