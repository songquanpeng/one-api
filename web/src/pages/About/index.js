import React, { useEffect, useState } from 'react';
import { Header, Segment } from 'semantic-ui-react';
import { API, showError } from '../../helpers';
import { marked } from 'marked';

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
      setAbout('Failed to load about content...');
    }
    setAboutLoaded(true);
  };

  useEffect(() => {
    displayAbout().then();
  }, []);

  return (
    <>
      {
        aboutLoaded && about === '' ? <>
          <Segment>
            <Header as='h3'>About</Header>
            <p>About content can be set on the settings page and supports HTML & Markdown</p>
            Repository address：
            <a href='https://github.com/analogpvt/one-api'>
              https://github.com/analogpvt/one-api
            </a>
          </Segment>
        </> : <>
          {
            about.startsWith('https://') ? <iframe
              src={about}
              style={{ width: '100%', height: '100vh', border: 'none' }}
            /> : <Segment>
              <div style={{ fontSize: 'larger' }} dangerouslySetInnerHTML={{ __html: about }}></div>
            </Segment>
          }
        </>
      }
    </>
  );
};


export default About;
