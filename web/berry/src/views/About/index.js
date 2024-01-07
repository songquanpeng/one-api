import React, { useEffect, useState } from 'react';
import { API } from 'utils/api';
import { showError } from 'utils/common';
import { marked } from 'marked';
import { Box, Container, Typography } from '@mui/material';
import MainCard from 'ui-component/cards/MainCard';

const About = () => {
  const [about, setAbout] = useState('');
  const [aboutLoaded, setAboutLoaded] = useState(false);

  const displayAbout = async () => {
    setAbout(localStorage.getItem('about') || '');
    const res = await API.get('/api/about');
    const { success, message, data } = res.data;
    if (success) {
      let aboutContent = data;
      if (!data.startsWith('https://')) {
        aboutContent = marked.parse(data);
      }
      setAbout(aboutContent);
      localStorage.setItem('about', aboutContent);
    } else {
      showError(message);
      setAbout('加载关于内容失败...');
    }
    setAboutLoaded(true);
  };

  useEffect(() => {
    displayAbout().then();
  }, []);

  return (
    <>
      {aboutLoaded && about === '' ? (
        <>
          <Box>
            <Container sx={{ paddingTop: '40px' }}>
              <MainCard title="关于">
                <Typography variant="body2">
                  可在设置页面设置关于内容，支持 HTML & Markdown <br />
                  项目仓库地址：
                  <a href="https://github.com/songquanpeng/one-api">https://github.com/songquanpeng/one-api</a>
                </Typography>
              </MainCard>
            </Container>
          </Box>
        </>
      ) : (
        <>
          <Box>
            {about.startsWith('https://') ? (
              <iframe title="about" src={about} style={{ width: '100%', height: '100vh', border: 'none' }} />
            ) : (
              <>
                <Container>
                  <div style={{ fontSize: 'larger' }} dangerouslySetInnerHTML={{ __html: about }}></div>
                </Container>
              </>
            )}
          </Box>
        </>
      )}
    </>
  );
};

export default About;
