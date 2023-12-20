import PropTypes from 'prop-types';
import * as Yup from 'yup';
import { Formik } from 'formik';
import { useTheme } from '@mui/material/styles';
import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Divider,
  FormControl,
  InputLabel,
  OutlinedInput,
  InputAdornment,
  Select,
  MenuItem,
  IconButton,
  FormHelperText
} from '@mui/material';

import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';

import { renderQuotaWithPrompt, showSuccess, showError } from 'utils/common';
import { API } from 'utils/api';

const validationSchema = Yup.object().shape({
  is_edit: Yup.boolean(),
  username: Yup.string().required('用户名 不能为空'),
  display_name: Yup.string(),
  password: Yup.string().when('is_edit', {
    is: false,
    then: Yup.string().required('密码 不能为空'),
    otherwise: Yup.string()
  }),
  group: Yup.string().when('is_edit', {
    is: false,
    then: Yup.string().required('用户组 不能为空'),
    otherwise: Yup.string()
  }),
  quota: Yup.number().when('is_edit', {
    is: false,
    then: Yup.number().min(0, '额度 不能小于 0'),
    otherwise: Yup.number()
  })
});

const originInputs = {
  is_edit: false,
  username: '',
  display_name: '',
  password: '',
  group: 'default',
  quota: 0
};

const EditModal = ({ open, userId, onCancel, onOk }) => {
  const theme = useTheme();
  const [inputs, setInputs] = useState(originInputs);
  const [groupOptions, setGroupOptions] = useState([]);
  const [showPassword, setShowPassword] = useState(false);

  const submit = async (values, { setErrors, setStatus, setSubmitting }) => {
    setSubmitting(true);

    let res;
    if (values.is_edit) {
      res = await API.put(`/api/user/`, { ...values, id: parseInt(userId) });
    } else {
      res = await API.post(`/api/user/`, values);
    }
    const { success, message } = res.data;
    if (success) {
      if (values.is_edit) {
        showSuccess('用户更新成功！');
      } else {
        showSuccess('用户创建成功！');
      }
      setSubmitting(false);
      setStatus({ success: true });
      onOk(true);
    } else {
      showError(message);
      setErrors({ submit: message });
    }
  };

  const handleClickShowPassword = () => {
    setShowPassword(!showPassword);
  };

  const handleMouseDownPassword = (event) => {
    event.preventDefault();
  };

  const loadUser = async () => {
    let res = await API.get(`/api/user/${userId}`);
    const { success, message, data } = res.data;
    if (success) {
      data.is_edit = true;
      setInputs(data);
    } else {
      showError(message);
    }
  };

  const fetchGroups = async () => {
    try {
      let res = await API.get(`/api/group/`);
      setGroupOptions(res.data.data);
    } catch (error) {
      showError(error.message);
    }
  };

  useEffect(() => {
    fetchGroups().then();
    if (userId) {
      loadUser().then();
    } else {
      setInputs(originInputs);
    }
  }, [userId]);

  return (
    <Dialog open={open} onClose={onCancel} fullWidth maxWidth={'md'}>
      <DialogTitle sx={{ margin: '0px', fontWeight: 700, lineHeight: '1.55556', padding: '24px', fontSize: '1.125rem' }}>
        {userId ? '编辑用户' : '新建用户'}
      </DialogTitle>
      <Divider />
      <DialogContent>
        <Formik initialValues={inputs} enableReinitialize validationSchema={validationSchema} onSubmit={submit}>
          {({ errors, handleBlur, handleChange, handleSubmit, touched, values, isSubmitting }) => (
            <form noValidate onSubmit={handleSubmit}>
              <FormControl fullWidth error={Boolean(touched.username && errors.username)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="channel-username-label">用户名</InputLabel>
                <OutlinedInput
                  id="channel-username-label"
                  label="用户名"
                  type="text"
                  value={values.username}
                  name="username"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{ autoComplete: 'username' }}
                  aria-describedby="helper-text-channel-username-label"
                />
                {touched.username && errors.username && (
                  <FormHelperText error id="helper-tex-channel-username-label">
                    {errors.username}
                  </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth error={Boolean(touched.display_name && errors.display_name)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="channel-display_name-label">显示名称</InputLabel>
                <OutlinedInput
                  id="channel-display_name-label"
                  label="显示名称"
                  type="text"
                  value={values.display_name}
                  name="display_name"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{ autoComplete: 'display_name' }}
                  aria-describedby="helper-text-channel-display_name-label"
                />
                {touched.display_name && errors.display_name && (
                  <FormHelperText error id="helper-tex-channel-display_name-label">
                    {errors.display_name}
                  </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth error={Boolean(touched.password && errors.password)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="channel-password-label">密码</InputLabel>
                <OutlinedInput
                  id="channel-password-label"
                  label="密码"
                  type={showPassword ? 'text' : 'password'}
                  value={values.password}
                  name="password"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{ autoComplete: 'password' }}
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
                  aria-describedby="helper-text-channel-password-label"
                />
                {touched.password && errors.password && (
                  <FormHelperText error id="helper-tex-channel-password-label">
                    {errors.password}
                  </FormHelperText>
                )}
              </FormControl>

              {values.is_edit && (
                <>
                  <FormControl fullWidth error={Boolean(touched.quota && errors.quota)} sx={{ ...theme.typography.otherInput }}>
                    <InputLabel htmlFor="channel-quota-label">额度</InputLabel>
                    <OutlinedInput
                      id="channel-quota-label"
                      label="额度"
                      type="number"
                      value={values.quota}
                      name="quota"
                      endAdornment={<InputAdornment position="end">{renderQuotaWithPrompt(values.quota)}</InputAdornment>}
                      onBlur={handleBlur}
                      onChange={handleChange}
                      aria-describedby="helper-text-channel-quota-label"
                      disabled={values.unlimited_quota}
                    />

                    {touched.quota && errors.quota && (
                      <FormHelperText error id="helper-tex-channel-quota-label">
                        {errors.quota}
                      </FormHelperText>
                    )}
                  </FormControl>

                  <FormControl fullWidth error={Boolean(touched.group && errors.group)} sx={{ ...theme.typography.otherInput }}>
                    <InputLabel htmlFor="channel-group-label">分组</InputLabel>
                    <Select
                      id="channel-group-label"
                      label="分组"
                      value={values.group}
                      name="group"
                      onBlur={handleBlur}
                      onChange={handleChange}
                      MenuProps={{
                        PaperProps: {
                          style: {
                            maxHeight: 200
                          }
                        }
                      }}
                    >
                      {groupOptions.map((option) => {
                        return (
                          <MenuItem key={option} value={option}>
                            {option}
                          </MenuItem>
                        );
                      })}
                    </Select>
                    {touched.group && errors.group && (
                      <FormHelperText error id="helper-tex-channel-group-label">
                        {errors.group}
                      </FormHelperText>
                    )}
                  </FormControl>
                </>
              )}
              <DialogActions>
                <Button onClick={onCancel}>取消</Button>
                <Button disableElevation disabled={isSubmitting} type="submit" variant="contained" color="primary">
                  提交
                </Button>
              </DialogActions>
            </form>
          )}
        </Formik>
      </DialogContent>
    </Dialog>
  );
};

export default EditModal;

EditModal.propTypes = {
  open: PropTypes.bool,
  userId: PropTypes.number,
  onCancel: PropTypes.func,
  onOk: PropTypes.func
};
