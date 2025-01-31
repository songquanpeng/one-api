import React, { useEffect, useState } from 'react';
import { Card } from 'semantic-ui-react';
import { API, showError } from '../../helpers';
import { marked } from 'marked';

const About = () => {
  const [about, setAbout] = useState('');
  const [aboutLoaded, setAboutLoaded] = useState(false);

  // ... 其他函数保持不变 ...

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>关于系统</Card.Header>
          {aboutLoaded && about === '' ? (
            <>
              <p>可在设置页面设置关于内容，支持 HTML & Markdown</p>
              项目仓库地址：
              <a href='https://github.com/songquanpeng/one-api'>
                https://github.com/songquanpeng/one-api
              </a>
            </>
          ) : (
            <>
              {about.startsWith('https://') ? (
                <iframe
                  src={about}
                  style={{ width: '100%', height: '100vh', border: 'none' }}
                />
              ) : (
                <div
                  style={{ fontSize: 'larger' }}
                  dangerouslySetInnerHTML={{ __html: about }}
                ></div>
              )}
            </>
          )}
        </Card.Content>
      </Card>
    </div>
  );
};

export default About;
