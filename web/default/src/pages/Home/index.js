import React, { useContext, useEffect, useState } from 'react';
import { Card, Grid, Header, Segment } from 'semantic-ui-react';
import { API, showError, showNotice, timestamp2string } from '../../helpers';
import { StatusContext } from '../../context/Status';
import { marked } from 'marked';
import { UserContext } from '../../context/User';
import { Link } from 'react-router-dom';

const Home = () => {
  const [statusState, statusDispatch] = useContext(StatusContext);
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');
  const [userState] = useContext(UserContext);

  const displayNotice = async () => {
    const res = await API.get('/api/notice');
    const { success, message, data } = res.data;
    if (success) {
      let oldNotice = localStorage.getItem('notice');
      if (data !== oldNotice && data !== '') {
        const htmlNotice = marked(data);
        showNotice(htmlNotice, true);
        localStorage.setItem('notice', data);
      }
    } else {
      showError(message);
    }
  };

  const displayHomePageContent = async () => {
    setHomePageContent(localStorage.getItem('home_page_content') || '');
    const res = await API.get('/api/home_page_content');
    const { success, message, data } = res.data;
    if (success) {
      let content = data;
      if (!data.startsWith('https://')) {
        content = marked.parse(data);
      }
      setHomePageContent(content);
      localStorage.setItem('home_page_content', content);
    } else {
      showError(message);
      setHomePageContent('加载首页内容失败...');
    }
    setHomePageContentLoaded(true);
  };

  const getStartTimeString = () => {
    const timestamp = statusState?.status?.start_time;
    return timestamp2string(timestamp);
  };

  useEffect(() => {
    displayNotice().then();
    displayHomePageContent().then();
  }, []);

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>欢迎使用 One API</Card.Header>
          <Card.Description style={{ lineHeight: '1.6' }}>
            <p>
              One API 是一个 OpenAI
              接口管理和分发系统，可以帮助您更好地管理和使用 OpenAI 的 API。
            </p>
            {!userState.user && (
              <p>
                如需使用，请先<Link to='/login'>登录</Link>或
                <Link to='/register'>注册</Link>。
              </p>
            )}
          </Card.Description>
        </Card.Content>
      </Card>

      {homePageContentLoaded && homePageContent === '' ? (
        <Card fluid className='chart-card'>
          <Card.Content>
            <Card.Header>
              <Header as='h3'>系统状况</Header>
            </Card.Header>
            <Grid columns={2} stackable>
              <Grid.Column>
                <Card
                  fluid
                  className='chart-card'
                  style={{ boxShadow: '0 1px 3px rgba(0,0,0,0.12)' }}
                >
                  <Card.Content>
                    <Card.Header>
                      <Header as='h3' style={{ color: '#444' }}>
                        系统信息
                      </Header>
                    </Card.Header>
                    <Card.Description
                      style={{ lineHeight: '2', marginTop: '1em' }}
                    >
                      <p
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '0.5em',
                        }}
                      >
                        <i className='info circle icon'></i>
                        <span style={{ fontWeight: 'bold' }}>名称：</span>
                        <span>{statusState?.status?.system_name}</span>
                      </p>
                      <p
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '0.5em',
                        }}
                      >
                        <i className='code branch icon'></i>
                        <span style={{ fontWeight: 'bold' }}>版本：</span>
                        <span>{statusState?.status?.version || 'unknown'}</span>
                      </p>
                      <p
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '0.5em',
                        }}
                      >
                        <i className='github icon'></i>
                        <span style={{ fontWeight: 'bold' }}>源码：</span>
                        <a
                          href='https://github.com/songquanpeng/one-api'
                          target='_blank'
                          style={{ color: '#2185d0' }}
                        >
                          GitHub 仓库
                        </a>
                      </p>
                      <p
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '0.5em',
                        }}
                      >
                        <i className='clock outline icon'></i>
                        <span style={{ fontWeight: 'bold' }}>启动时间：</span>
                        <span>{getStartTimeString()}</span>
                      </p>
                    </Card.Description>
                  </Card.Content>
                </Card>
              </Grid.Column>

              <Grid.Column>
                <Card
                  fluid
                  className='chart-card'
                  style={{ boxShadow: '0 1px 3px rgba(0,0,0,0.12)' }}
                >
                  <Card.Content>
                    <Card.Header>
                      <Header as='h3' style={{ color: '#444' }}>
                        系统配置
                      </Header>
                    </Card.Header>
                    <Card.Description
                      style={{ lineHeight: '2', marginTop: '1em' }}
                    >
                      <p
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '0.5em',
                        }}
                      >
                        <i className='envelope icon'></i>
                        <span style={{ fontWeight: 'bold' }}>邮箱验证：</span>
                        <span
                          style={{
                            color: statusState?.status?.email_verification
                              ? '#21ba45'
                              : '#db2828',
                            fontWeight: '500',
                          }}
                        >
                          {statusState?.status?.email_verification
                            ? '已启用'
                            : '未启用'}
                        </span>
                      </p>
                      <p
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '0.5em',
                        }}
                      >
                        <i className='github icon'></i>
                        <span style={{ fontWeight: 'bold' }}>
                          GitHub 身份验证：
                        </span>
                        <span
                          style={{
                            color: statusState?.status?.github_oauth
                              ? '#21ba45'
                              : '#db2828',
                            fontWeight: '500',
                          }}
                        >
                          {statusState?.status?.github_oauth
                            ? '已启用'
                            : '未启用'}
                        </span>
                      </p>
                      <p
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '0.5em',
                        }}
                      >
                        <i className='wechat icon'></i>
                        <span style={{ fontWeight: 'bold' }}>
                          微信身份验证：
                        </span>
                        <span
                          style={{
                            color: statusState?.status?.wechat_login
                              ? '#21ba45'
                              : '#db2828',
                            fontWeight: '500',
                          }}
                        >
                          {statusState?.status?.wechat_login
                            ? '已启用'
                            : '未启用'}
                        </span>
                      </p>
                      <p
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '0.5em',
                        }}
                      >
                        <i className='shield alternate icon'></i>
                        <span style={{ fontWeight: 'bold' }}>
                          Turnstile 校验：
                        </span>
                        <span
                          style={{
                            color: statusState?.status?.turnstile_check
                              ? '#21ba45'
                              : '#db2828',
                            fontWeight: '500',
                          }}
                        >
                          {statusState?.status?.turnstile_check
                            ? '已启用'
                            : '未启用'}
                        </span>
                      </p>
                    </Card.Description>
                  </Card.Content>
                </Card>
              </Grid.Column>
            </Grid>
          </Card.Content>
        </Card>
      ) : (
        <>
          {homePageContent.startsWith('https://') ? (
            <iframe
              src={homePageContent}
              style={{ width: '100%', height: '100vh', border: 'none' }}
            />
          ) : (
            <div
              style={{ fontSize: 'larger' }}
              dangerouslySetInnerHTML={{ __html: homePageContent }}
            ></div>
          )}
        </>
      )}
    </div>
  );
};

export default Home;
