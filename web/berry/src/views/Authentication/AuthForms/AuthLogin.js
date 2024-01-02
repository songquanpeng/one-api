import { useState } from 'react';
import { useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

// material-ui
import { useTheme } from '@mui/material/styles';
import {
  Box,
  Button,
  Divider,
  FormControl,
  FormHelperText,
  Grid,
  IconButton,
  InputAdornment,
  InputLabel,
  OutlinedInput,
  Stack,
  Typography,
  useMediaQuery
} from '@mui/material';

// third party
import * as Yup from 'yup';
import { Formik } from 'formik';

// project imports
import useLogin from 'hooks/useLogin';
import AnimateButton from 'ui-component/extended/AnimateButton';
import WechatModal from 'views/Authentication/AuthForms/WechatModal';

// assets
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';

import Github from 'assets/images/icons/github.svg';
import Wechat from 'assets/images/icons/wechat.svg';
import { onGitHubOAuthClicked } from 'utils/common';

// ============================|| FIREBASE - LOGIN ||============================ //

const LoginForm = ({ ...others }) => {
  const theme = useTheme();
  const { login, wechatLogin } = useLogin();
  const [openWechat, setOpenWechat] = useState(false);
  const matchDownSM = useMediaQuery(theme.breakpoints.down('md'));
  const customization = useSelector((state) => state.customization);
  const siteInfo = useSelector((state) => state.siteInfo);
  // const [checked, setChecked] = useState(true);

  let tripartiteLogin = false;
  if (siteInfo.github_oauth || siteInfo.wechat_login) {
    tripartiteLogin = true;
  }

  const handleWechatOpen = () => {
    setOpenWechat(true);
  };

  const handleWechatClose = () => {
    setOpenWechat(false);
  };

  const [showPassword, setShowPassword] = useState(false);
  const handleClickShowPassword = () => {
    setShowPassword(!showPassword);
  };

  const handleMouseDownPassword = (event) => {
    event.preventDefault();
  };

  return (
    <>
      {tripartiteLogin && (
        <Grid container direction="column" justifyContent="center" spacing={2}>
          {siteInfo.github_oauth && (
            <Grid item xs={12}>
              <AnimateButton>
                <Button
                  disableElevation
                  fullWidth
                  onClick={() => onGitHubOAuthClicked(siteInfo.github_client_id)}
                  size="large"
                  variant="outlined"
                  sx={{
                    color: 'grey.700',
                    backgroundColor: theme.palette.grey[50],
                    borderColor: theme.palette.grey[100]
                  }}
                >
                  <Box sx={{ mr: { xs: 1, sm: 2, width: 20 }, display: 'flex', alignItems: 'center' }}>
                    <img src={Github} alt="github" width={25} height={25} style={{ marginRight: matchDownSM ? 8 : 16 }} />
                  </Box>
                  使用 Github 登录
                </Button>
              </AnimateButton>
            </Grid>
          )}
          {siteInfo.wechat_login && (
            <Grid item xs={12}>
              <AnimateButton>
                <Button
                  disableElevation
                  fullWidth
                  onClick={handleWechatOpen}
                  size="large"
                  variant="outlined"
                  sx={{
                    color: 'grey.700',
                    backgroundColor: theme.palette.grey[50],
                    borderColor: theme.palette.grey[100]
                  }}
                >
                  <Box sx={{ mr: { xs: 1, sm: 2, width: 20 }, display: 'flex', alignItems: 'center' }}>
                    <img src={Wechat} alt="Wechat" width={25} height={25} style={{ marginRight: matchDownSM ? 8 : 16 }} />
                  </Box>
                  使用 Wechat 登录
                </Button>
              </AnimateButton>
              <WechatModal open={openWechat} handleClose={handleWechatClose} wechatLogin={wechatLogin} qrCode={siteInfo.wechat_qrcode} />
            </Grid>
          )}
          <Grid item xs={12}>
            <Box
              sx={{
                alignItems: 'center',
                display: 'flex'
              }}
            >
              <Divider sx={{ flexGrow: 1 }} orientation="horizontal" />

              <Button
                variant="outlined"
                sx={{
                  cursor: 'unset',
                  m: 2,
                  py: 0.5,
                  px: 7,
                  borderColor: `${theme.palette.grey[100]} !important`,
                  color: `${theme.palette.grey[900]}!important`,
                  fontWeight: 500,
                  borderRadius: `${customization.borderRadius}px`
                }}
                disableRipple
                disabled
              >
                OR
              </Button>

              <Divider sx={{ flexGrow: 1 }} orientation="horizontal" />
            </Box>
          </Grid>
        </Grid>
      )}

      <Formik
        initialValues={{
          username: '',
          password: '',
          submit: null
        }}
        validationSchema={Yup.object().shape({
          username: Yup.string().max(255).required('Username is required'),
          password: Yup.string().max(255).required('Password is required')
        })}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          const { success, message } = await login(values.username, values.password);
          if (success) {
            setStatus({ success: true });
          } else {
            setStatus({ success: false });
            if (message) {
              setErrors({ submit: message });
            }
          }
          setSubmitting(false);
        }}
      >
        {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, touched, values }) => (
          <form noValidate onSubmit={handleSubmit} {...others}>
            <FormControl fullWidth error={Boolean(touched.username && errors.username)} sx={{ ...theme.typography.customInput }}>
              <InputLabel htmlFor="outlined-adornment-username-login">用户名</InputLabel>
              <OutlinedInput
                id="outlined-adornment-username-login"
                type="text"
                value={values.username}
                name="username"
                onBlur={handleBlur}
                onChange={handleChange}
                label="用户名"
                inputProps={{ autoComplete: 'username' }}
              />
              {touched.username && errors.username && (
                <FormHelperText error id="standard-weight-helper-text-username-login">
                  {errors.username}
                </FormHelperText>
              )}
            </FormControl>

            <FormControl fullWidth error={Boolean(touched.password && errors.password)} sx={{ ...theme.typography.customInput }}>
              <InputLabel htmlFor="outlined-adornment-password-login">密码</InputLabel>
              <OutlinedInput
                id="outlined-adornment-password-login"
                type={showPassword ? 'text' : 'password'}
                value={values.password}
                name="password"
                onBlur={handleBlur}
                onChange={handleChange}
                endAdornment={
                  <InputAdornment position="end">
                    <IconButton
                      aria-label="toggle password visibility"
                      onClick={handleClickShowPassword}
                      onMouseDown={handleMouseDownPassword}
                      edge="end"
                      size="large"
                    >
                      {showPassword ? <Visibility /> : <VisibilityOff />}
                    </IconButton>
                  </InputAdornment>
                }
                label="Password"
              />
              {touched.password && errors.password && (
                <FormHelperText error id="standard-weight-helper-text-password-login">
                  {errors.password}
                </FormHelperText>
              )}
            </FormControl>
            <Stack direction="row" alignItems="center" justifyContent="space-between" spacing={1}>
              {/* <FormControlLabel
                control={
                  <Checkbox checked={checked} onChange={(event) => setChecked(event.target.checked)} name="checked" color="primary" />
                }
                label="记住我"
              /> */}
              <Typography
                component={Link}
                to="/reset"
                variant="subtitle1"
                color="primary"
                sx={{ textDecoration: 'none', cursor: 'pointer' }}
              >
                忘记密码?
              </Typography>
            </Stack>
            {errors.submit && (
              <Box sx={{ mt: 3 }}>
                <FormHelperText error>{errors.submit}</FormHelperText>
              </Box>
            )}

            <Box sx={{ mt: 2 }}>
              <AnimateButton>
                <Button disableElevation disabled={isSubmitting} fullWidth size="large" type="submit" variant="contained" color="primary">
                  登录
                </Button>
              </AnimateButton>
            </Box>
          </form>
        )}
      </Formik>
    </>
  );
};

export default LoginForm;
