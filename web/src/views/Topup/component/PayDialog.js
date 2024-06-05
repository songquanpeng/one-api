import PropTypes from 'prop-types';
import { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogTitle, IconButton, Stack, Typography } from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import { useTheme } from '@mui/material/styles';
import { QRCode } from 'react-qrcode-logo';
import successSvg from 'assets/images/success.svg';
import { API } from 'utils/api';
import { showError } from 'utils/common';

const PayDialog = ({ open, onClose, amount, uuid }) => {
  const theme = useTheme();
  const defaultLogo = theme.palette.mode === 'light' ? '/logo-loading.svg' : '/logo-loading-white.svg';
  const [message, setMessage] = useState('正在拉起支付中...');
  const [loading, setLoading] = useState(false);
  const [qrCodeUrl, setQrCodeUrl] = useState(null);
  const [success, setSuccess] = useState(false);
  const [intervalId, setIntervalId] = useState(null);

  useEffect(() => {
    if (!open) {
      return;
    }
    setMessage('正在拉起支付中...');
    setLoading(true);

    API.post('/api/user/order', {
      uuid: uuid,
      amount: Number(amount)
    }).then((response) => {
      if (!response.data.success) {
        showError(response.data.message);
        setLoading(false);
        onClose();
        return;
      }

      const { type, data } = response.data.data;
      if (type === 1) {
        setMessage('等待支付中...');
        const form = document.createElement('form');
        form.method = data.method;
        form.action = data.url;
        form.target = '_blank';
        for (const key in data.params) {
          const input = document.createElement('input');
          input.name = key;
          input.value = data.params[key];
          form.appendChild(input);
        }
        document.body.appendChild(form);
        form.submit();
        document.body.removeChild(form);
      } else if (type === 2) {
        setQrCodeUrl(data.url);
        setLoading(false);
        setMessage('请扫码支付');
      }
      pollOrderStatus(response.data.data.trade_no);
    });
  }, [open, onClose, amount, uuid]);

  const pollOrderStatus = (tradeNo) => {
    const id = setInterval(() => {
      API.get(`/api/user/order/status?trade_no=${tradeNo}`).then((response) => {
        if (response.data.success) {
          setMessage('支付成功');
          setLoading(false);
          setSuccess(true);
          clearInterval(id);
          setIntervalId(null);
        }
      });
    }, 3000);
    setIntervalId(id);
  };

  return (
    <Dialog open={open} fullWidth maxWidth={'sm'} disableEscapeKeyDown>
      <DialogTitle sx={{ margin: '0px', fontWeight: 700, lineHeight: '1.55556', padding: '24px', fontSize: '1.125rem' }}>支付</DialogTitle>
      <IconButton
        aria-label="close"
        onClick={() => {
          if (intervalId) {
            clearInterval(intervalId);
            setIntervalId(null);
          }
          onClose();
        }}
        sx={{
          position: 'absolute',
          right: 8,
          top: 8,
          color: (theme) => theme.palette.grey[500]
        }}
      >
        <CloseIcon />
      </IconButton>
      <DialogContent>
        <DialogContent>
          <Stack direction="column" justifyContent="center" alignItems="center" spacing={2}>
            {loading && <img src={defaultLogo} alt="loading" height="100" />}
            {qrCodeUrl && (
              <QRCode
                value={qrCodeUrl}
                size={256}
                qrStyle="dots"
                eyeRadius={20}
                fgColor={theme.palette.primary.main}
                bgColor={theme.palette.background.paper}
              />
            )}
            {success && <img src={successSvg} alt="success" height="100" />}
            <Typography variant="h3">{message}</Typography>
          </Stack>
        </DialogContent>
      </DialogContent>
    </Dialog>
  );
};

export default PayDialog;

PayDialog.propTypes = {
  open: PropTypes.bool,
  onClose: PropTypes.func,
  amount: PropTypes.number,
  uuid: PropTypes.string
};
