import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import { showError } from 'utils/common';
import useLogin from 'hooks/useLogin';

// material-ui
import { useTheme } from '@mui/material/styles';
import { Grid, Stack, Typography, useMediaQuery, CircularProgress } from '@mui/material';

// project imports
import AuthWrapper from '../AuthWrapper';
import AuthCardWrapper from '../AuthCardWrapper';
import Logo from 'ui-component/Logo';

// assets

// ================================|| AUTH3 - LOGIN ||================================ //

const GitHubOAuth = () => {
  const theme = useTheme();
  const matchDownSM = useMediaQuery(theme.breakpoints.down('md'));

  const [searchParams] = useSearchParams();
  const [prompt, setPrompt] = useState('处理中...');
  const { githubLogin } = useLogin();

  let navigate = useNavigate();

  const sendCode = async (code, state, count) => {
    const { success, message } = await githubLogin(code, state);
    if (!success) {
      if (message) {
        showError(message);
      }
      if (count === 0) {
        setPrompt(`操作失败，重定向至登录界面中...`);
        await new Promise((resolve) => setTimeout(resolve, 2000));
        navigate('/login');
        return;
      }
      count++;
      setPrompt(`出现错误，第 ${count} 次重试中...`);
      await new Promise((resolve) => setTimeout(resolve, 2000));
      await sendCode(code, state, count);
    }
  };

  useEffect(() => {
    let code = searchParams.get('code');
    let state = searchParams.get('state');
    sendCode(code, state, 0).then();
  }, []);

  return (
    <AuthWrapper>
      <Grid container direction="column" justifyContent="flex-end">
        <Grid item xs={12}>
          <Grid container justifyContent="center" alignItems="center" sx={{ minHeight: 'calc(100vh - 136px)' }}>
            <Grid item sx={{ m: { xs: 1, sm: 3 }, mb: 0 }}>
              <AuthCardWrapper>
                <Grid container spacing={2} alignItems="center" justifyContent="center">
                  <Grid item sx={{ mb: 3 }}>
                    <Link to="#">
                      <Logo />
                    </Link>
                  </Grid>
                  <Grid item xs={12}>
                    <Grid container direction={matchDownSM ? 'column-reverse' : 'row'} alignItems="center" justifyContent="center">
                      <Grid item>
                        <Stack alignItems="center" justifyContent="center" spacing={1}>
                          <Typography color={theme.palette.primary.main} gutterBottom variant={matchDownSM ? 'h3' : 'h2'}>
                            GitHub 登录
                          </Typography>
                        </Stack>
                      </Grid>
                    </Grid>
                  </Grid>
                  <Grid item xs={12} container direction="column" justifyContent="center" alignItems="center" style={{ height: '200px' }}>
                    <CircularProgress />
                    <Typography variant="h3" paddingTop={'20px'}>
                      {prompt}
                    </Typography>
                  </Grid>
                </Grid>
              </AuthCardWrapper>
            </Grid>
          </Grid>
        </Grid>
      </Grid>
    </AuthWrapper>
  );
};

export default GitHubOAuth;
