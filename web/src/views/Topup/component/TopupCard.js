import {
  Typography,
  Stack,
  OutlinedInput,
  InputAdornment,
  Button,
  InputLabel,
  FormControl,
  useMediaQuery,
  TextField,
  Box,
  Grid,
  Divider
} from '@mui/material';
import { IconBuildingBank } from '@tabler/icons-react';
import { useTheme } from '@mui/material/styles';
import SubCard from 'ui-component/cards/SubCard';
import UserCard from 'ui-component/cards/UserCard';
import AnimateButton from 'ui-component/extended/AnimateButton';
import { useSelector } from 'react-redux';
import PayDialog from './PayDialog';

import { API } from 'utils/api';
import React, { useEffect, useState } from 'react';
import { showError, showInfo, showSuccess, renderQuota, trims } from 'utils/common';

const TopupCard = () => {
  const theme = useTheme();
  const [redemptionCode, setRedemptionCode] = useState('');
  const [userQuota, setUserQuota] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [payment, setPayment] = useState([]);
  const [selectedPayment, setSelectedPayment] = useState(null);
  const [amount, setAmount] = useState(0);
  const [open, setOpen] = useState(false);

  const matchDownSM = useMediaQuery(theme.breakpoints.down('md'));
  const siteInfo = useSelector((state) => state.siteInfo);

  const topUp = async () => {
    if (redemptionCode === '') {
      showInfo('请输入充值码！');
      return;
    }
    setIsSubmitting(true);
    try {
      const res = await API.post('/api/user/topup', {
        key: trims(redemptionCode)
      });
      const { success, message, data } = res.data;
      if (success) {
        showSuccess('充值成功！');
        setUserQuota((quota) => {
          return quota + data;
        });
        setRedemptionCode('');
      } else {
        showError(message);
      }
    } catch (err) {
      showError('请求失败');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handlePay = () => {
    if (!selectedPayment) {
      showError('请选择支付方式');
      return;
    }

    if (amount <= 0 || amount < siteInfo.PaymentMinAmount) {
      showError('金额不能小于' + siteInfo.PaymentMinAmount);
      return;
    }

    if (amount > 1000000) {
      showError('金额不能大于1000000');
      return;
    }

    // 判读金额是否是正整数
    if (!/^[1-9]\d*$/.test(amount)) {
      showError('请输入正整数金额');
      return;
    }

    setOpen(true);
  };

  const getPayment = async () => {
    try {
      let res = await API.get(`/api/user/payment`);
      const { success, data } = res.data;
      if (success) {
        if (data.length > 0) {
          data.sort((a, b) => b.sort - a.sort);
          setPayment(data);
          setSelectedPayment(data[0]);
        }
      }
    } catch (error) {
      return;
    }
  };

  const openTopUpLink = () => {
    if (!siteInfo.top_up_link) {
      showError('超级管理员未设置充值链接！');
      return;
    }
    window.open(siteInfo.top_up_link, '_blank');
  };

  const getUserQuota = async () => {
    try {
      let res = await API.get(`/api/user/self`);
      const { success, message, data } = res.data;
      if (success) {
        setUserQuota(data.quota);
      } else {
        showError(message);
      }
    } catch (error) {
      return;
    }
  };

  const handlePaymentSelect = (payment) => {
    setSelectedPayment(payment);
  };

  const handleAmountChange = (event) => {
    const value = event.target.value;
    setAmount(value);
  };
  const calculateFee = () => {
    if (!selectedPayment) return 0;

    if (selectedPayment.fixed_fee > 0) {
      return Number(selectedPayment.fixed_fee);
    }

    return parseFloat(selectedPayment.percent_fee * Number(amount)).toFixed(2);
  };

  const calculateTotal = () => {
    if (amount === 0) return 0;

    let total = Number(amount) + Number(calculateFee());
    if (selectedPayment && selectedPayment.currency === 'CNY') {
      total = parseFloat((total * siteInfo.PaymentUSDRate).toFixed(2));
    }
    return total;
  };

  useEffect(() => {
    getPayment().then();
    getUserQuota().then();
  }, []);

  return (
    <UserCard>
      <Stack direction="row" alignItems="center" justifyContent="center" spacing={2} paddingTop={'20px'}>
        <IconBuildingBank color={theme.palette.primary.main} />
        <Typography variant="h4">当前额度:</Typography>
        <Typography variant="h4">{renderQuota(userQuota)}</Typography>
      </Stack>

      {payment.length > 0 && (
        <SubCard
          sx={{
            marginTop: '40px'
          }}
          title="在线充值"
        >
          <Stack spacing={2}>
            {payment.map((item, index) => (
              <AnimateButton key={index}>
                <Button
                  disableElevation
                  fullWidth
                  size="large"
                  variant="outlined"
                  onClick={() => handlePaymentSelect(item)}
                  sx={{
                    ...theme.typography.LoginButton,
                    border: selectedPayment === item ? `1px solid ${theme.palette.primary.main}` : '1px solid transparent'
                  }}
                >
                  <Box sx={{ mr: { xs: 1, sm: 2, width: 20 }, display: 'flex', alignItems: 'center' }}>
                    <img src={item.icon} alt="github" width={25} height={25} style={{ marginRight: matchDownSM ? 8 : 16 }} />
                  </Box>
                  {item.name}
                </Button>
              </AnimateButton>
            ))}
            <TextField label="金额" type="number" onChange={handleAmountChange} value={amount} />
            <Divider />
            <Grid container direction="row" justifyContent="flex-end" spacing={2}>
              <Grid item xs={9}>
                <Typography variant="h6" style={{ textAlign: 'right', fontSize: '0.875rem' }}>
                  充值金额:{' '}
                </Typography>
              </Grid>
              <Grid item xs={3}>
                ${Number(amount)}
              </Grid>
              {selectedPayment && (selectedPayment.percent_fee > 0 || selectedPayment.fixed_fee > 0) && (
                <>
                  <Grid item xs={9}>
                    <Typography variant="h6" style={{ textAlign: 'right', fontSize: '0.875rem' }}>
                      手续费:
                      {selectedPayment &&
                        (selectedPayment.fixed_fee > 0
                          ? '(固定)'
                          : selectedPayment.percent_fee > 0
                            ? `(${selectedPayment.percent_fee * 100}%)`
                            : '')}{' '}
                    </Typography>
                  </Grid>
                  <Grid item xs={3}>
                    ${calculateFee()}
                  </Grid>
                </>
              )}

              <Grid item xs={9}>
                <Typography variant="h6" style={{ textAlign: 'right', fontSize: '0.875rem' }}>
                  实际支付金额:{' '}
                </Typography>
              </Grid>
              <Grid item xs={3}>
                {calculateTotal()}{' '}
                {selectedPayment &&
                  (selectedPayment.currency === 'CNY' ? `CNY (汇率：${siteInfo.PaymentUSDRate})` : selectedPayment.currency)}
              </Grid>
            </Grid>
            <Divider />
            <Button variant="contained" onClick={handlePay}>
              充值
            </Button>
          </Stack>
          <PayDialog open={open} onClose={() => setOpen(false)} amount={amount} uuid={selectedPayment.uuid} />
        </SubCard>
      )}

      <SubCard
        sx={{
          marginTop: '40px'
        }}
        title="兑换码充值"
      >
        <FormControl fullWidth variant="outlined">
          <InputLabel htmlFor="key">兑换码</InputLabel>
          <OutlinedInput
            id="key"
            label="兑换码"
            type="text"
            value={redemptionCode}
            onChange={(e) => {
              setRedemptionCode(e.target.value);
            }}
            name="key"
            placeholder="请输入兑换码"
            endAdornment={
              <InputAdornment position="end">
                <Button variant="contained" onClick={topUp} disabled={isSubmitting}>
                  {isSubmitting ? '兑换中...' : '兑换'}
                </Button>
              </InputAdornment>
            }
            aria-describedby="helper-text-channel-quota-label"
          />
        </FormControl>

        <Stack justifyContent="center" alignItems={'center'} spacing={3} paddingTop={'20px'}>
          <Typography variant={'h4'} color={theme.palette.grey[700]}>
            还没有兑换码？ 点击获取兑换码：
          </Typography>
          <Button variant="contained" onClick={openTopUpLink}>
            获取兑换码
          </Button>
        </Stack>
      </SubCard>
    </UserCard>
  );
};

export default TopupCard;
