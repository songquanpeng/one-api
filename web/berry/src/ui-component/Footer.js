// material-ui
import { Link, Container, Box } from '@mui/material';
import React from 'react';
import { useSelector } from 'react-redux';

// ==============================|| FOOTER - AUTHENTICATION 2 & 3 ||============================== //

const Footer = () => {
  const siteInfo = useSelector((state) => state.siteInfo);

  return (
    <Container sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '64px' }}>
      <Box sx={{ textAlign: 'center' }}>
        {siteInfo.footer_html ? (
          <div className="custom-footer" dangerouslySetInnerHTML={{ __html: siteInfo.footer_html }}></div>
        ) : (
          <>
            <Link href="https://github.com/songquanpeng/one-api" target="_blank">
              {siteInfo.system_name} {process.env.REACT_APP_VERSION}{' '}
            </Link>
            由{' '}
            <Link href="https://github.com/songquanpeng" target="_blank">
              JustSong
            </Link>{' '}
            构建，主题 berry 来自{' '}
            <Link href="https://github.com/MartialBE" target="_blank">
              MartialBE
            </Link>{' '}，源代码遵循
            <Link href="https://opensource.org/licenses/mit-license.php"> MIT 协议</Link>
          </>
        )}
      </Box>
    </Container>
  );
};

export default Footer;
