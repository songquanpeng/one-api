import { useState, useEffect } from 'react';
import UserCard from 'ui-component/cards/UserCard';
import {
  Card,
  Button,
  InputLabel,
  FormControl,
  OutlinedInput,
  Stack,
  Alert,
  Divider,
  Chip,
  Typography,
  SvgIcon,
  useMediaQuery
} from '@mui/material';
import Grid from '@mui/material/Unstable_Grid2';
import SubCard from 'ui-component/cards/SubCard';
import { IconBrandWechat, IconBrandGithub, IconMail, IconBrandTelegram } from '@tabler/icons-react';
import Label from 'ui-component/Label';
import { API } from 'utils/api';
import { showError, showSuccess, onGitHubOAuthClicked, copy, trims, onLarkOAuthClicked } from 'utils/common';
import * as Yup from 'yup';
import WechatModal from 'views/Authentication/AuthForms/WechatModal';
import { useSelector } from 'react-redux';
import EmailModal from './component/EmailModal';
import Turnstile from 'react-turnstile';
import { ReactComponent as Lark } from 'assets/images/icons/lark.svg';
import { useTheme } from '@mui/material/styles';

const validationSchema = Yup.object().shape({
  username: Yup.string().required('用户名 不能为空').min(3, '用户名 不能小于 3 个字符'),
  display_name: Yup.string(),
  password: Yup.string().test('password', '密码不能小于 8 个字符', (val) => {
    return !val || val.length >= 8;
  })
});

export default function Profile() {
  const [inputs, setInputs] = useState([]);
  const [turnstileEnabled, setTurnstileEnabled] = useState(false);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [turnstileToken, setTurnstileToken] = useState('');
  const [openWechat, setOpenWechat] = useState(false);
  const [openEmail, setOpenEmail] = useState(false);
  const status = useSelector((state) => state.siteInfo);
  const theme = useTheme();
  const matchDownSM = useMediaQuery(theme.breakpoints.down('md'));

  const handleWechatOpen = () => {
    setOpenWechat(true);
  };

  const handleWechatClose = () => {
    setOpenWechat(false);
  };

  const handleInputChange = (event) => {
    let { name, value } = event.target;
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const loadUser = async () => {
    try {
      let res = await API.get(`/api/user/self`);
      const { success, message, data } = res.data;
      if (success) {
        setInputs(data);
      } else {
        showError(message);
      }
    } catch (error) {
      return;
    }
  };

  const bindWeChat = async (code) => {
    if (code === '') return;
    try {
      const res = await API.get(`/api/oauth/wechat/bind?code=${code}`);
      const { success, message } = res.data;
      if (success) {
        showSuccess('微信账户绑定成功！');
      }
      return { success, message };
    } catch (err) {
      // 请求失败，设置错误信息
      return { success: false, message: '' };
    }
  };

  const generateAccessToken = async () => {
    try {
      const res = await API.get('/api/user/token');
      const { success, message, data } = res.data;
      if (success) {
        setInputs((inputs) => ({ ...inputs, access_token: data }));
        copy(data, '令牌');
      } else {
        showError(message);
      }
    } catch (error) {
      return;
    }
  };

  const submit = async () => {
    try {
      let inputValue = inputs;
      // inputValue.username = trims(inputValue.username);
      inputValue.display_name = trims(inputValue.display_name);
      await validationSchema.validate(inputValue);
      const res = await API.put(`/api/user/self`, inputValue);
      const { success, message } = res.data;
      if (success) {
        showSuccess('用户信息更新成功！');
      } else {
        showError(message);
      }
    } catch (err) {
      showError(err.message);
    }
  };

  useEffect(() => {
    if (status) {
      if (status.turnstile_check) {
        setTurnstileEnabled(true);
        setTurnstileSiteKey(status.turnstile_site_key);
      }
    }
    loadUser().then();
  }, [status]);

  return (
    <>
      <UserCard>
        <Card sx={{ paddingTop: '20px' }}>
          <Stack spacing={2}>
            <Stack
              direction={matchDownSM ? 'column' : 'row'}
              alignItems="center"
              justifyContent="center"
              spacing={2}
              sx={{ paddingBottom: '20px' }}
            >
              <Label variant="ghost" color={inputs.wechat_id ? 'primary' : 'default'}>
                <IconBrandWechat /> {inputs.wechat_id || '未绑定'}
              </Label>
              <Label variant="ghost" color={inputs.github_id ? 'primary' : 'default'}>
                <IconBrandGithub /> {inputs.github_id || '未绑定'}
              </Label>
              <Label variant="ghost" color={inputs.email ? 'primary' : 'default'}>
                <IconMail /> {inputs.email || '未绑定'}
              </Label>
              <Label variant="ghost" color={inputs.telegram_id ? 'primary' : 'default'}>
                <IconBrandTelegram /> {inputs.telegram_id || '未绑定'}
              </Label>
              <Label variant="ghost" color={inputs.lark_id ? 'primary' : 'default'}>
                <SvgIcon component={Lark} inheritViewBox="0 0 24 24" /> {inputs.lark_id || '未绑定'}
              </Label>
            </Stack>
            <SubCard title="个人信息">
              <Grid container spacing={2}>
                <Grid xs={12}>
                  <FormControl fullWidth variant="outlined">
                    <InputLabel htmlFor="username">用户名</InputLabel>
                    <OutlinedInput
                      id="username"
                      label="用户名"
                      type="text"
                      value={inputs.username || ''}
                      // onChange={handleInputChange}
                      disabled
                      name="username"
                      placeholder="请输入用户名"
                    />
                  </FormControl>
                </Grid>
                <Grid xs={12}>
                  <FormControl fullWidth variant="outlined">
                    <InputLabel htmlFor="password">密码</InputLabel>
                    <OutlinedInput
                      id="password"
                      label="密码"
                      type="password"
                      value={inputs.password || ''}
                      onChange={handleInputChange}
                      name="password"
                      placeholder="请输入密码"
                    />
                  </FormControl>
                </Grid>
                <Grid xs={12}>
                  <FormControl fullWidth variant="outlined">
                    <InputLabel htmlFor="display_name">显示名称</InputLabel>
                    <OutlinedInput
                      id="display_name"
                      label="显示名称"
                      type="text"
                      value={inputs.display_name || ''}
                      onChange={handleInputChange}
                      name="display_name"
                      placeholder="请输入显示名称"
                    />
                  </FormControl>
                </Grid>
                <Grid xs={12}>
                  <Button variant="contained" color="primary" onClick={submit}>
                    提交
                  </Button>
                </Grid>
              </Grid>
            </SubCard>
            <SubCard title="账号绑定">
              <Grid container spacing={2}>
                {status.wechat_login && !inputs.wechat_id && (
                  <Grid xs={12} md={4}>
                    <Button variant="contained" onClick={handleWechatOpen}>
                      绑定微信账号
                    </Button>
                  </Grid>
                )}
                {status.github_oauth && !inputs.github_id && (
                  <Grid xs={12} md={4}>
                    <Button variant="contained" onClick={() => onGitHubOAuthClicked(status.github_client_id, true)}>
                      绑定GitHub账号
                    </Button>
                  </Grid>
                )}

                {status.lark_client_id && !inputs.lark_id && (
                  <Grid xs={12} md={4}>
                    <Button variant="contained" onClick={() => onLarkOAuthClicked(status.lark_client_id)}>
                      绑定 飞书 账号
                    </Button>
                  </Grid>
                )}

                <Grid xs={12} md={4}>
                  <Button
                    variant="contained"
                    onClick={() => {
                      setOpenEmail(true);
                    }}
                  >
                    {inputs.email ? '更换邮箱' : '绑定邮箱'}
                  </Button>
                  {turnstileEnabled ? (
                    <Turnstile
                      sitekey={turnstileSiteKey}
                      onVerify={(token) => {
                        setTurnstileToken(token);
                      }}
                    />
                  ) : (
                    <></>
                  )}
                </Grid>

                {status.telegram_bot && ( //&& !inputs.telegram_id
                  <Grid xs={12} md={12}>
                    <Stack spacing={2}>
                      <Divider />

                      <Alert severity="info">
                        <Typography variant="h3">Telegram 机器人</Typography>
                        <br />
                        <Typography variant="body1">
                          1. 点击下方按钮，将会在 Telegram 中打开 机器人，点击 /start 开始。
                          <br />
                          <Chip
                            icon={<IconBrandTelegram />}
                            label={'@' + status.telegram_bot}
                            color="primary"
                            variant="outlined"
                            size="small"
                            onClick={() => window.open('https://t.me/' + status.telegram_bot, '_blank')}
                          />
                          <br />
                          <br />
                          2. 向机器人发送/bind命令后，输入下方的访问令牌即可绑定。(如果没有生成，请点击下方按钮生成)
                        </Typography>
                      </Alert>
                      {/* <Typography variant="">  */}
                    </Stack>
                  </Grid>
                )}
              </Grid>
            </SubCard>
            <SubCard title="其他">
              <Grid container spacing={2}>
                <Grid xs={12}>
                  <Alert severity="info">注意，此处生成的令牌用于系统管理，而非用于请求 OpenAI 相关的服务，请知悉。</Alert>
                </Grid>
                {inputs.access_token && (
                  <Grid xs={12}>
                    <Alert severity="error">
                      你的访问令牌是: <b>{inputs.access_token}</b> <br />
                      请妥善保管。如有泄漏，请立即重置。
                    </Alert>
                  </Grid>
                )}
                <Grid xs={12}>
                  <Button variant="contained" onClick={generateAccessToken}>
                    {inputs.access_token ? '重置访问令牌' : '生成访问令牌'}
                  </Button>
                </Grid>
              </Grid>
            </SubCard>
          </Stack>
        </Card>
      </UserCard>
      <WechatModal open={openWechat} handleClose={handleWechatClose} wechatLogin={bindWeChat} qrCode={status.wechat_qrcode} />
      <EmailModal
        open={openEmail}
        turnstileToken={turnstileToken}
        turnstileEnabled={turnstileEnabled}
        handleClose={() => {
          setOpenEmail(false);
        }}
      />
    </>
  );
}
