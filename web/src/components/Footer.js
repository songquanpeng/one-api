import React, { useEffect, useState } from 'react';

import { Container, Segment } from 'semantic-ui-react';

const Footer = () => {
  const [Footer, setFooter] = useState('');
  useEffect(() => {
    let savedFooter = localStorage.getItem('footer_html');
    if (!savedFooter) savedFooter = '';
    setFooter(savedFooter);
  });

  return (
    <Segment vertical>
      <Container textAlign="center">
        {Footer === '' ? (
          <div className="custom-footer">
            <a
              href="https://github.com/songquanpeng/gin-template"
              target="_blank"
            >
              项目模板 {process.env.REACT_APP_VERSION}{' '}
            </a>
            由{' '}
            <a href="https://github.com/songquanpeng" target="_blank">
              JustSong
            </a>{' '}
            构建，源代码遵循{' '}
            <a href="https://opensource.org/licenses/mit-license.php">
              MIT 协议
            </a>
          </div>
        ) : (
          <div
            className="custom-footer"
            dangerouslySetInnerHTML={{ __html: Footer }}
          ></div>
        )}
      </Container>
    </Segment>
  );
};

export default Footer;
