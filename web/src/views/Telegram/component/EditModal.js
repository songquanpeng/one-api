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
  FormHelperText,
  Select,
  MenuItem,
  TextField
} from '@mui/material';

import { showSuccess, showError } from 'utils/common';
import { API } from 'utils/api';

const validationSchema = Yup.object().shape({
  is_edit: Yup.boolean(),
  command: Yup.string().required('命令 不能为空'),
  description: Yup.string().required('说明 不能为空'),
  parse_mode: Yup.string().required('消息类型 不能为空'),
  reply_message: Yup.string().required('消息内容 不能为空')
});

const originInputs = {
  command: '',
  description: '',
  parse_mode: 'MarkdownV2',
  reply_message: ''
};

const EditModal = ({ open, actionId, onCancel, onOk }) => {
  const theme = useTheme();
  const [inputs, setInputs] = useState(originInputs);

  const submit = async (values, { setErrors, setStatus, setSubmitting }) => {
    setSubmitting(true);

    let res;
    try {
      if (values.is_edit) {
        res = await API.post(`/api/option/telegram/`, { ...values, id: parseInt(actionId) });
      } else {
        res = await API.post(`/api/option/telegram/`, values);
      }
      const { success, message } = res.data;
      if (success) {
        if (values.is_edit) {
          showSuccess('菜单更新成功！');
        } else {
          showSuccess('菜单创建成功！');
        }
        setSubmitting(false);
        setStatus({ success: true });
        onOk(true);
      } else {
        showError(message);
        setErrors({ submit: message });
      }
    } catch (error) {
      return;
    }
  };

  const load = async () => {
    try {
      let res = await API.get(`/api/option/telegram/${actionId}`);
      const { success, message, data } = res.data;
      if (success) {
        data.is_edit = true;
        setInputs(data);
      } else {
        showError(message);
      }
    } catch (error) {
      return;
    }
  };

  useEffect(() => {
    if (actionId) {
      load().then();
    } else {
      setInputs(originInputs);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [actionId]);

  return (
    <Dialog open={open} onClose={onCancel} fullWidth maxWidth={'md'}>
      <DialogTitle sx={{ margin: '0px', fontWeight: 700, lineHeight: '1.55556', padding: '24px', fontSize: '1.125rem' }}>
        {actionId ? '编辑菜单' : '新建菜单'}
      </DialogTitle>
      <Divider />
      <DialogContent>
        <Formik initialValues={inputs} enableReinitialize validationSchema={validationSchema} onSubmit={submit}>
          {({ errors, handleBlur, handleChange, handleSubmit, touched, values, isSubmitting }) => (
            <form noValidate onSubmit={handleSubmit}>
              <FormControl fullWidth error={Boolean(touched.command && errors.command)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="channel-command-label">命令</InputLabel>
                <OutlinedInput
                  id="channel-command-label"
                  label="命令"
                  type="text"
                  value={values.command}
                  name="command"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{ autoComplete: 'command' }}
                  aria-describedby="helper-text-channel-command-label"
                />
                {touched.command && errors.command && (
                  <FormHelperText error id="helper-tex-channel-command-label">
                    {errors.command}
                  </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth error={Boolean(touched.description && errors.description)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="channel-description-label">说明</InputLabel>
                <OutlinedInput
                  id="channel-description-label"
                  label="说明"
                  type="text"
                  value={values.description}
                  name="description"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{ autoComplete: 'description' }}
                  aria-describedby="helper-text-channel-description-label"
                />
                {touched.description && errors.description && (
                  <FormHelperText error id="helper-tex-channel-description-label">
                    {errors.description}
                  </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth error={Boolean(touched.parse_mode && errors.parse_mode)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="channel-parse_mode-label">消息类型</InputLabel>
                <Select
                  id="channel-parse_mode-label"
                  label="消息类型"
                  value={values.parse_mode}
                  name="parse_mode"
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
                  <MenuItem key="MarkdownV2" value="MarkdownV2">
                    {' '}
                    MarkdownV2{' '}
                  </MenuItem>
                  <MenuItem key="Markdown" value="Markdown">
                    {' '}
                    Markdown{' '}
                  </MenuItem>
                  <MenuItem key="html" value="html">
                    {' '}
                    html{' '}
                  </MenuItem>
                </Select>
                {touched.parse_mode && errors.parse_mode && (
                  <FormHelperText error id="helper-tex-channel-parse_mode-label">
                    {errors.parse_mode}
                  </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth error={Boolean(touched.reply_message && errors.reply_message)} sx={{ ...theme.typography.otherInput }}>
                <TextField
                  multiline
                  id="channel-reply_message-label"
                  label="消息内容"
                  value={values.reply_message}
                  name="reply_message"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  aria-describedby="helper-text-channel-reply_message-label"
                  minRows={5}
                  placeholder="消息内容"
                />
                {touched.reply_message && errors.reply_message && (
                  <FormHelperText error id="helper-tex-channel-reply_message-label">
                    {errors.reply_message}
                  </FormHelperText>
                )}
              </FormControl>

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
  actionId: PropTypes.number,
  onCancel: PropTypes.func,
  onOk: PropTypes.func
};
