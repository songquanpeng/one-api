import React, { useEffect, useState } from 'react';
import { Header, Segment } from 'semantic-ui-react';
import { API, showError } from '../../helpers';
import { marked } from 'marked';

const About = () => {
  const [about, setAbout] = useState('');

  const displayAbout = async () => {
    const res = await API.get('/api/about');
    const { success, message, data } = res.data;
    if (success) {
      let HTMLAbout = marked.parse(data);
      localStorage.setItem('about', HTMLAbout);
      setAbout(HTMLAbout);
    } else {
      showError(message);
      setAbout('加载关于内容失败...');
    }
  };

  useEffect(() => {
    displayAbout().then();
  }, []);

  return (
    <>
      <Segment>
        {
          about === '' ? <>
            <Header as='h3'>关于</Header>
            <p>可在设置页面设置关于内容，支持 HTML & Markdown</p>
            项目仓库地址：
            <a href="https://github.com/songquanpeng/gin-template">
              https://github.com/songquanpeng/gin-template
            </a>
          </> : <>
            <div dangerouslySetInnerHTML={{ __html: about}}></div>
          </>
        }
      </Segment>
    </>
  );
};


export default About;
