import React, { useEffect, useState } from 'react';
import {
  Button,
  Form,
  Grid,
  Header,
  Card,
  Statistic,
  Divider,
} from 'semantic-ui-react';
import { API, showError, showInfo, showSuccess } from '../../helpers';
import { renderQuota } from '../../helpers/render';

const TopUp = () => {
  const [redemptionCode, setRedemptionCode] = useState('');
  const [topUpLink, setTopUpLink] = useState('');
  const [userQuota, setUserQuota] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [user, setUser] = useState({});

  const topUp = async () => {
    if (redemptionCode === '') {
      showInfo('请输入充值码！');
      return;
    }
    setIsSubmitting(true);
    try {
      const res = await API.post('/api/user/topup', {
        key: redemptionCode,
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
    } catch (err) {
      showError('请求失败');
    } finally {
      setIsSubmitting(false);
    }
  };

  const openTopUpLink = () => {
    if (!topUpLink) {
      showError('超级管理员未设置充值链接！');
      return;
    }
    let url = new URL(topUpLink);
    let username = user.username;
    let user_id = user.id;
    // add username and user_id to the topup link
    url.searchParams.append('username', username);
    url.searchParams.append('user_id', user_id);
    url.searchParams.append('transaction_id', crypto.randomUUID());
    window.open(url.toString(), '_blank');
  };

  const getUserQuota = async () => {
    let res = await API.get(`/api/user/self`);
    const { success, message, data } = res.data;
    if (success) {
      setUserQuota(data.quota);
      setUser(data);
    } else {
      showError(message);
    }
  };

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
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header>
            <Header as='h2' style={{ marginBottom: '1.5em' }}>
              <i
                className='money bill alternate outline icon'
                style={{ color: '#2185d0' }}
              ></i>
              充值中心
            </Header>
          </Card.Header>

          <Grid columns={2} stackable>
            <Grid.Column>
              <Card
                fluid
                style={{
                  height: '100%',
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                }}
              >
                <Card.Content>
                  <Card.Header>
                    <Header
                      as='h3'
                      style={{ color: '#2185d0', marginBottom: '1em' }}
                    >
                      <i className='shopping cart icon'></i>
                      在线充值
                    </Header>
                  </Card.Header>
                  <Card.Description>
                    <div style={{ textAlign: 'center', padding: '1em 0' }}>
                      <Statistic>
                        <Statistic.Value style={{ color: '#2185d0' }}>
                          {renderQuota(userQuota)}
                        </Statistic.Value>
                        <Statistic.Label>当前可用额度</Statistic.Label>
                      </Statistic>

                      <Button
                        primary
                        size='large'
                        onClick={openTopUpLink}
                        style={{ marginTop: '2em', width: '80%' }}
                      >
                        <i className='credit card icon'></i>
                        立即充值
                      </Button>
                    </div>
                  </Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>

            <Grid.Column>
              <Card
                fluid
                style={{
                  height: '100%',
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                }}
              >
                <Card.Content>
                  <Card.Header>
                    <Header
                      as='h3'
                      style={{ color: '#21ba45', marginBottom: '1em' }}
                    >
                      <i className='ticket alternate icon'></i>
                      兑换码充值
                    </Header>
                  </Card.Header>
                  <Card.Description>
                    <Form size='large'>
                      <Form.Input
                        fluid
                        icon='key'
                        iconPosition='left'
                        placeholder='请输入兑换码'
                        value={redemptionCode}
                        onChange={(e) => {
                          setRedemptionCode(e.target.value);
                        }}
                        onPaste={(e) => {
                          e.preventDefault();
                          const pastedText = e.clipboardData.getData('text');
                          setRedemptionCode(pastedText.trim());
                        }}
                        action={
                          <Button
                            icon='paste'
                            content='粘贴'
                            onClick={async () => {
                              try {
                                const text =
                                  await navigator.clipboard.readText();
                                setRedemptionCode(text.trim());
                              } catch (err) {
                                showError('无法访问剪贴板，请手动粘贴');
                              }
                            }}
                          />
                        }
                        style={{ marginBottom: '1em' }}
                      />
                      <Button
                        color='green'
                        fluid
                        size='large'
                        onClick={topUp}
                        loading={isSubmitting}
                        disabled={isSubmitting}
                      >
                        {isSubmitting ? '兑换中...' : '立即兑换'}
                      </Button>
                    </Form>
                  </Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>
          </Grid>
        </Card.Content>
      </Card>
    </div>
  );
};

export default TopUp;
