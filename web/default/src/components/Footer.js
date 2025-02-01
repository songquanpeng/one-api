import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Container, Segment } from 'semantic-ui-react';
import { getFooterHTML, getSystemName } from '../helpers';

const Footer = () => {
  const { t } = useTranslation();
  const systemName = getSystemName();
  const [footer, setFooter] = useState(getFooterHTML());
  let remainCheckTimes = 5;

  const loadFooter = () => {
    let footer_html = localStorage.getItem('footer_html');
    if (footer_html) {
      setFooter(footer_html);
    }
  };

  useEffect(() => {
    const timer = setInterval(() => {
      if (remainCheckTimes <= 0) {
        clearInterval(timer);
        return;
      }
      remainCheckTimes--;
      loadFooter();
    }, 200);
    return () => clearTimeout(timer);
  }, []);

  return (
    <Segment vertical>
      <Container textAlign='center' style={{ color: '#666666' }}>
        {footer ? (
          <div
            className='custom-footer'
            dangerouslySetInnerHTML={{ __html: footer }}
          ></div>
        ) : (
          <div className='custom-footer'>
            <a href='https://github.com/songquanpeng/one-api' target='_blank'>
              {systemName} {process.env.REACT_APP_VERSION}{' '}
            </a>
            {t('footer.built_by')}{' '}
            <a href='https://github.com/songquanpeng' target='_blank'>
              {t('footer.built_by_name')}
            </a>{' '}
            {t('footer.license')}{' '}
            <a href='https://opensource.org/licenses/mit-license.php'>
              {t('footer.mit')}
            </a>
          </div>
        )}
      </Container>
    </Segment>
  );
};

export default Footer;
