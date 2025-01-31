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
          <Card.Header className='header'>
            欢迎使用 One API
          </Card.Header>
          <Card.Description style={{ lineHeight: '1.6' }}>
            <p>
              One API 是一个 OpenAI 接口管理和分发系统，可以帮助您更好地管理和使用 OpenAI 的 API。
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

      <Grid columns={3} stackable className='charts-grid'>
        <Grid.Column>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header className='header'>
                使用说明
              </Card.Header>
              <Card.Description style={{ lineHeight: '1.6' }}>
                <p>1. 登录并获取令牌</p>
                <p>2. 在您的应用中使用令牌</p>
                <p>3. 监控使用情况和费用</p>
              </Card.Description>
            </Card.Content>
          </Card>
        </Grid.Column>

        <Grid.Column>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header className='header'>
                功能特点
              </Card.Header>
              <Card.Description style={{ lineHeight: '1.6' }}>
                <p>• 多渠道接口管理</p>
                <p>• 实时监控和统计</p>
                <p>• 灵活的配额控制</p>
              </Card.Description>
            </Card.Content>
          </Card>
        </Grid.Column>

        <Grid.Column>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header className='header'>
                技术支持
              </Card.Header>
              <Card.Description style={{ lineHeight: '1.6' }}>
                <p>• 完整的API文档</p>
                <p>• 详细的使用教程</p>
                <p>• 及时的问题解答</p>
              </Card.Description>
            </Card.Content>
          </Card>
        </Grid.Column>
      </Grid>

      {homePageContentLoaded && homePageContent === '' ? (
        <>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header className='header'>系统状况</Card.Header>
              <Grid columns={2} stackable>
                <Grid.Column>
                  <Card fluid className='chart-card'>
                    <Card.Content>
                      <Card.Header className='header'>系统信息</Card.Header>
                      <Card.Meta>系统信息总览</Card.Meta>
                      <Card.Description style={{ lineHeight: '1.6' }}>
                        <p>名称：{statusState?.status?.system_name}</p>
                        <p>版本：{statusState?.status?.version ? statusState?.status?.version : "unknown"}</p>
                        <p>
                          源码：
                          <a href='https://github.com/songquanpeng/one-api' target='_blank'>
                            https://github.com/songquanpeng/one-api
                          </a>
                        </p>
                        <p>启动时间：{getStartTimeString()}</p>
                      </Card.Description>
                    </Card.Content>
                  </Card>
                </Grid.Column>
                <Grid.Column>
                  <Card fluid className='chart-card'>
                    <Card.Content>
                      <Card.Header className='header'>系统配置</Card.Header>
                      <Card.Meta>系统配置总览</Card.Meta>
                      <Card.Description style={{ lineHeight: '1.6' }}>
                        <p>
                          邮箱验证：
                          {statusState?.status?.email_verification === true
                            ? '已启用'
                            : '未启用'}
                        </p>
                        <p>
                          GitHub 身份验证：
                          {statusState?.status?.github_oauth === true
                            ? '已启用'
                            : '未启用'}
                        </p>
                        <p>
                          微信身份验证：
                          {statusState?.status?.wechat_login === true
                            ? '已启用'
                            : '未启用'}
                        </p>
                        <p>
                          Turnstile 用户校验：
                          {statusState?.status?.turnstile_check === true
                            ? '已启用'
                            : '未启用'}
                        </p>
                      </Card.Description>
                    </Card.Content>
                  </Card>
                </Grid.Column>
              </Grid>
            </Card.Content>
          </Card>
        </>
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
