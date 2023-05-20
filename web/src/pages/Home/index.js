import React, { useContext, useEffect, useState } from 'react';
import { Card, Grid, Header, Segment } from 'semantic-ui-react';
import { API, showError, showNotice, timestamp2string } from '../../helpers';
import { StatusContext } from '../../context/Status';
import { marked } from 'marked';

const Home = () => {
  const [statusState, statusDispatch] = useContext(StatusContext);
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');

  const displayNotice = async () => {
    const res = await API.get('/api/notice');
    const { success, message, data } = res.data;
    if (success) {
      let oldNotice = localStorage.getItem('notice');
      if (data !== oldNotice && data !== '') {
        showNotice(data);
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
      setHomePageContent('Failed to load home page content...');
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
    <>
      {
        homePageContentLoaded && homePageContent === '' ? <>
          <Segment>
            <Header as='h3'>System Status</Header>
            <Grid columns={2} stackable>
              <Grid.Column>
                <Card fluid>
                  <Card.Content>
                    <Card.Header>System Message</Card.Header>
                    <Card.Meta>System Information Overview</Card.Meta>
                    <Card.Description>
                      <p>Name: {statusState?.status?.system_name}</p>
                      <p>Version: {statusState?.status?.version}</p>
                      <p>
                        Source Code:
                        <a
                          href='https://github.com/songquanpeng/one-api'
                          target='_blank'
                        >
                          https://github.com/songquanpeng/one-api
                        </a>
                      </p>
                      <p>Start Time: {getStartTimeString()}</p>
                    </Card.Description>
                  </Card.Content>
                </Card>
              </Grid.Column>
              <Grid.Column>
                <Card fluid>
                  <Card.Content>
                    <Card.Header>System Configuration</Card.Header>
                    <Card.Meta>System configuration overview</Card.Meta>
                    <Card.Description>
                      <p>
                      E-mail verification:
                        {statusState?.status?.email_verification === true
                          ? 'Enabled'
                          : 'Not Enabled'}
                      </p>
                      <p>
                      GitHub authentication:
                        {statusState?.status?.github_oauth === true
                          ? 'Enabled'
                          : 'Not Enabled'}
                      </p>
                      <p>
                      WeChat authentication:
                        {statusState?.status?.wechat_login === true
                          ? 'Enabled'
                          : 'Not Enabled'}
                      </p>
                      <p>
                        Turnstile user validation:
                        {statusState?.status?.turnstile_check === true
                          ? 'Enabled'
                          : 'Not Enabled'}
                      </p>
                    </Card.Description>
                  </Card.Content>
                </Card>
              </Grid.Column>
            </Grid>
          </Segment>
        </> : <>
          {
            homePageContent.startsWith('https://') ? <iframe
              src={homePageContent}
              style={{ width: '100%', height: '100vh', border: 'none' }}
            /> : <div style={{ fontSize: 'larger' }} dangerouslySetInnerHTML={{ __html: homePageContent }}></div>
          }
        </>
      }

    </>
  );
};

export default Home;
