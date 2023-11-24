import React, { useEffect, useState } from 'react';
import { Button, Form, Grid, Header, Segment, Statistic } from 'semantic-ui-react';
import { API, showError, showInfo, showSuccess } from '../../helpers';
import { renderQuota } from '../../helpers/render';

const TopUp = () => {
  const [redemptionCode, setRedemptionCode] = useState('');
  const [topUpLink, setTopUpLink] = useState('');
  const [userQuota, setUserQuota] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const topUp = async () => {
    if (redemptionCode === '') {
      showInfo('请输入充值码！')
      return;
    }
    setIsSubmitting(true);
    try {
      const res = await API.post('/api/user/topup', {
        key: redemptionCode
      });
      const { success, message, data, upgradedToVIP } = res.data;
      if (success) {
        if (upgradedToVIP) {  // 如果用户成功升级为 VIP
          showSuccess('充值成功，升级为 VIP 会员');
        } else {
          showSuccess('充值成功');
        }
        setUserQuota((quota) => {
          const newQuota = quota + data;
          return newQuota;
        });
        setRedemptionCode('');
      } else {
        showError(message);
      }
    } catch (err) {
      showError('失败，请右下角联系客服');
    } finally {
      setIsSubmitting(false);
    }
  };

  const openTopUpLink = () => {
    if (!topUpLink) {
      showError('超级管理员未设置充值链接！');
      return;
    }
    window.open(topUpLink, '_blank');
  };

  const getUserQuota = async () => {
    let res = await API.get(`/api/user/self`);
    const { success, message, data } = res.data;
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
            <Button
              negative
              icon='shop'
              labelPosition='left'
              content='获取兑换码'
              style={{ backgroundColor: 'var(--czl-blue-700)' }}
              onClick={openTopUpLink}
            />
            <Button
              negative
              icon='exchange'
              labelPosition='left'
              content={isSubmitting ? '兑换中...' : '兑换'}
              style={{ backgroundColor: '#FFFFFF00',color: 'var(--czl-blue-700)',border: '1px solid var(--czl-blue-200)' }}
              onClick={topUp}
              disabled={isSubmitting}
            />

          </Form>
        </Grid.Column>
        <Grid.Column>
          <Statistic.Group widths='one'>
            <Statistic>
              <Statistic.Value style={{ color: 'var(--czl-blue-800)' }}>{renderQuota(userQuota)}</Statistic.Value>
              <Statistic.Label style={{ color: 'var(--czl-blue-800)' }}>剩余额度</Statistic.Label>
            </Statistic>
          </Statistic.Group>
        </Grid.Column>
      </Grid>
    </Segment>
  );
};

export default TopUp;