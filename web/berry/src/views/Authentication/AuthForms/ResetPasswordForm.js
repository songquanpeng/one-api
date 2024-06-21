import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';

// material-ui
import { Button, Stack, Typography, Alert } from '@mui/material';

// assets
import { showError, copy } from 'utils/common';
import { API } from 'utils/api';

// ===========================|| FIREBASE - REGISTER ||=========================== //

const ResetPasswordForm = () => {
  const [searchParams] = useSearchParams();
  const [inputs, setInputs] = useState({
    email: '',
    token: ''
  });
  const [newPassword, setNewPassword] = useState('');

  const submit = async () => {
    const res = await API.post(`/api/user/reset`, inputs);
    const { success, message } = res.data;
    if (success) {
      let password = res.data.data;
      setNewPassword(password);
      copy(password, '新密码');
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    let email = searchParams.get('email');
    let token = searchParams.get('token');
    setInputs({
      token,
      email
    });
  }, []);

  return (
    <Stack spacing={3} padding={'24px'} justifyContent={'center'} alignItems={'center'}>
      {!inputs.email || !inputs.token ? (
        <Typography variant="h3" sx={{ textDecoration: 'none' }}>
          无效的链接
        </Typography>
      ) : newPassword ? (
        <Alert severity="error">
          你的新密码是: <b>{newPassword}</b> <br />
          请登录后及时修改密码
        </Alert>
      ) : (
        <Button fullWidth onClick={submit} size="large" type="submit" variant="contained" color="primary">
          点击重置密码
        </Button>
      )}
    </Stack>
  );
};

export default ResetPasswordForm;
