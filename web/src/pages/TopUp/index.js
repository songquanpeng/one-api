import React, { useEffect, useState } from 'react';
import { Button, Form, Grid, Header, Segment, Statistic } from 'semantic-ui-react';
import { API, showError, showInfo, showSuccess } from '../../helpers';
import { renderQuota } from '../../helpers/render';

const TopUp = () => {
  const [redemptionCode, setRedemptionCode] = useState('');
  const [topUpLink, setTopUpLink] = useState('');
  const [userQuota, setUserQuota] = useState(0);
  const [userGroup, setUserGroup] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // 升级用户组
  const updateUserGroupIfNecessary = async (quota) => {
    if (userGroup === 'vip') return; // 添加这一行

    if (quota >= 5*500000) {
      try {
        const res = await API.post('/api/user/manage', {
          username: localStorage.getItem('username'),
          newGroup: 'vip'
        });
        const { success, message } = res.data;
        if (success) {
          showSuccess('已成功升级为 VIP 会员！');
        } else {
          showError('请右下角联系客服');
        }
      } catch (err) {
        showError('请右下角联系客服');
      }
    }
  };

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
      const { success, message, data } = res.data;
      if (success) {
        showSuccess('充值成功！');
        setUserQuota((quota) => {
          const newQuota = quota + data;
          updateUserGroupIfNecessary(newQuota);
          return newQuota;
        });
        setRedemptionCode('');
      } else {
        showError(message);
      }
    } catch (err) {
      showError('请右下角联系客服');
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
      setUserGroup(data.group); // 添加这一行
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
            <Button negative style={{ backgroundColor: 'var(--czl-primary-color)' }} onClick={openTopUpLink}>
              获取兑换码
            </Button>
            <Button negative style={{ backgroundColor: 'var(--czl-success-color)' }} onClick={topUp} disabled={isSubmitting}>
              {isSubmitting ? '兑换中...' : '兑换'}
            </Button>


          </Form>
        </Grid.Column>
        <Grid.Column>
          <Statistic.Group widths='one'>
            <Statistic>
              <Statistic.Value style={{ color: 'var(--czl-error-color)' }}>{renderQuota(userQuota)}</Statistic.Value>
              <Statistic.Label style={{ color: 'var(--czl-error-color)' }}>剩余额度</Statistic.Label>
            </Statistic>
          </Statistic.Group>
        </Grid.Column>
      </Grid>
    </Segment>
  );
};

export default TopUp;