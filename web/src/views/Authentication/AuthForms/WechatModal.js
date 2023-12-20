// WechatModal.js
import PropTypes from 'prop-types';
import React from 'react';
import { Dialog, DialogTitle, DialogContent, TextField, Button, Typography, Grid } from '@mui/material';
import { Formik, Form, Field } from 'formik';
import { showError } from 'utils/common';
import * as Yup from 'yup';

const validationSchema = Yup.object().shape({
  code: Yup.string().required('验证码不能为空')
});

const WechatModal = ({ open, handleClose, wechatLogin, qrCode }) => {
  const handleSubmit = (values) => {
    const { success, message } = wechatLogin(values.code);
    if (success) {
      handleClose();
    } else {
      showError(message || '未知错误');
    }
  };

  return (
    <Dialog open={open} onClose={handleClose}>
      <DialogTitle>微信验证码登录</DialogTitle>
      <DialogContent>
        <Grid container direction="column" alignItems="center">
          <img src={qrCode} alt="二维码" style={{ maxWidth: '300px', maxHeight: '300px', width: 'auto', height: 'auto' }} />
          <Typography
            variant="body2"
            color="text.secondary"
            style={{ marginTop: '10px', textAlign: 'center', wordWrap: 'break-word', maxWidth: '300px' }}
          >
            请使用微信扫描二维码关注公众号，输入「验证码」获取验证码（三分钟内有效）
          </Typography>
          <Formik initialValues={{ code: '' }} validationSchema={validationSchema} onSubmit={handleSubmit}>
            {({ errors, touched }) => (
              <Form style={{ width: '100%' }}>
                <Grid item xs={12}>
                  <Field
                    as={TextField}
                    name="code"
                    label="验证码"
                    error={touched.code && Boolean(errors.code)}
                    helperText={touched.code && errors.code}
                    fullWidth
                  />
                </Grid>
                <Grid item xs={12}>
                  <Button type="submit" fullWidth>
                    提交
                  </Button>
                </Grid>
              </Form>
            )}
          </Formik>
        </Grid>
      </DialogContent>
    </Dialog>
  );
};

export default WechatModal;

WechatModal.propTypes = {
  open: PropTypes.bool,
  handleClose: PropTypes.func,
  wechatLogin: PropTypes.func,
  qrCode: PropTypes.string
};
