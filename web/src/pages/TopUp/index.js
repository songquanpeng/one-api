import React, { useEffect, useState } from 'react';
import { Button, Form, Grid, Header, Segment, Statistic } from 'semantic-ui-react';
import { API, showError, showSuccess } from '../../helpers';

const TopUp = () => {
  const [redemptionCode, setRedemptionCode] = useState('');
  const [topUpLink, setTopUpLink] = useState('');
  const [userQuota, setUserQuota] = useState(0);

  const topUp = async () => {
    if (redemptionCode === '') {
      return;
    }
    const res = await API.post('/api/user/topup', {
      key: redemptionCode
    });
    const { success, message, data } = res.data;
    if (success) {
      showSuccess('充值成功！');
      setUserQuota((quota) => {
        return quota + data;
      });
      setRedemptionCode('');
    } else {
      showError(message);
    }
  };

  const openTopUpLink = () => {
    if (!topUpLink) {
      showError('超级管理员未设置充值链接！');
      return;
    }
    window.open(topUpLink, '_blank');
  };

  const getUserQuota = async ()=>{
    let res  = await API.get(`/api/user/self`);
    const {success, message, data} = res.data;
    if (success) {
      setUserQuota(data.quota);
    } else {
      showError(message);
    }
  }

  useEffect(() => {
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      if (status.top_up_link) {
        setTopUpLink(status.top_up_link);
      }
    }
    getUserQuota().then();
  }, []);

  return (
    <Segment>
      <Header as='h3'>充值额度</Header>
      <Grid columns={2} stackable>
        <Grid.Column>
          <Form>
            <Form.Input
              placeholder='兑换码'
              name='redemptionCode'
              value={redemptionCode}
              onChange={(e) => {
                setRedemptionCode(e.target.value);
              }}
            />
            <Button color='green' onClick={openTopUpLink}>
              获取兑换码
            </Button>
            <Button color='yellow' onClick={topUp}>
              充值
            </Button>
          </Form>
        </Grid.Column>
        <Grid.Column>
          <Statistic.Group widths='one'>
            <Statistic>
              <Statistic.Value>{userQuota.toLocaleString()}</Statistic.Value>
              <Statistic.Label>剩余额度</Statistic.Label>
            </Statistic>
          </Statistic.Group>
        </Grid.Column>
      </Grid>
    </Segment>
  );
};


export default TopUp;
