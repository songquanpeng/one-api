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
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Divider,
  SvgIcon
} from '@mui/material';
import Grid from '@mui/material/Unstable_Grid2';
import SubCard from 'ui-component/cards/SubCard';
import { IconBrandWechat, IconBrandGithub, IconMail } from '@tabler/icons-react';
import Label from 'ui-component/Label';
import { API } from 'utils/api';
import { onOidcClicked, showError, showSuccess } from 'utils/common';
import { onGitHubOAuthClicked, onLarkOAuthClicked, copy } from 'utils/common';
import * as Yup from 'yup';
import WechatModal from 'views/Authentication/AuthForms/WechatModal';
import { useSelector } from 'react-redux';
import EmailModal from './component/EmailModal';
import Turnstile from 'react-turnstile';
import { ReactComponent as Lark } from 'assets/images/icons/lark.svg';
import { ReactComponent as OIDC } from 'assets/images/icons/oidc.svg';

const validationSchema = Yup.object().shape({
  username: Yup.string().required('用户名 不能为空').min(3, '用户名 不能小于 3 个字符'),
  display_name: Yup.string(),
  password: Yup.string().test('password', '密码不能小于 8 个字符', (val) => {
    return !val || val.length >= 8;
  })
});

export default function Profile() {
  const [inputs, setInputs] = useState([]);
  const [showAccountDeleteModal, setShowAccountDeleteModal] = useState(false);
  const [turnstileEnabled, setTurnstileEnabled] = useState(false);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [turnstileToken, setTurnstileToken] = useState('');
  const [openWechat, setOpenWechat] = useState(false);
  const [openEmail, setOpenEmail] = useState(false);
  const status = useSelector((state) => state.siteInfo);

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
    let res = await API.get(`/api/user/self`);
    const { success, message, data } = res.data;
    if (success) {
      setInputs(data);
    } else {
      showError(message);
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
    const res = await API.get('/api/user/token');
    const { success, message, data } = res.data;
    if (success) {
      setInputs((inputs) => ({ ...inputs, access_token: data }));
      copy(data, '访问令牌');
    } else {
      showError(message);
    }

    console.log(turnstileEnabled, turnstileSiteKey, status);
  };

  const submit = async () => {
    try {
      await validationSchema.validate(inputs);
      const res = await API.put(`/api/user/self`, inputs);
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

  function getOidcId(){
    if (!inputs.oidc_id) return '';
    let oidc_id = inputs.oidc_id;
    if (inputs.oidc_id.length > 8) {
      oidc_id = inputs.oidc_id.slice(0, 6) + '...' + inputs.oidc_id.slice(-6);
    }
    return oidc_id;
  }

  return (
    <>
      <UserCard>
        <Card sx={{ paddingTop: '20px' }}>
          <Stack spacing={2}>
            <Stack direction="row" alignItems="center" justifyContent="center" spacing={2} sx={{ paddingBottom: '20px' }}>
              <Label variant="ghost" color={inputs.wechat_id ? 'primary' : 'default'}>
                <IconBrandWechat /> {inputs.wechat_id || '未绑定'}
              </Label>
              <Label variant="ghost" color={inputs.github_id ? 'primary' : 'default'}>
                <IconBrandGithub /> {inputs.github_id || '未绑定'}
              </Label>
              <Label variant="ghost" color={inputs.email ? 'primary' : 'default'}>
                <IconMail /> {inputs.email || '未绑定'}
              </Label>
              <Label variant="ghost" color={inputs.lark_id ? 'primary' : 'default'}>
                <SvgIcon component={Lark} inheritViewBox="0 0 24 24" /> {inputs.lark_id || '未绑定'}
              </Label>
              <Label variant="ghost" color={inputs.oidc_id ? 'primary' : 'default'}>
                <SvgIcon component={OIDC} inheritViewBox="0 0 24 24" /> {getOidcId() || '未绑定'}
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
                      onChange={handleInputChange}
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
                      绑定 GitHub 账号
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
                {status.oidc && !inputs.oidc_id && (
                  <Grid xs={12} md={4}>
                    <Button variant="contained" onClick={() => onOidcClicked(status.oidc_authorization_endpoint,status.oidc_client_id,true)}>
                      绑定 OIDC 账号
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

                <Grid xs={12}>
                  <Button
                    variant="contained"
                    color="error"
                    onClick={() => {
                      setShowAccountDeleteModal(true);
                    }}
                  >
                    删除帐号
                  </Button>
                </Grid>
              </Grid>
            </SubCard>
          </Stack>
        </Card>
      </UserCard>
      <Dialog open={showAccountDeleteModal} onClose={() => setShowAccountDeleteModal(false)} maxWidth={'md'}>
        <DialogTitle sx={{ margin: '0px', fontWeight: 500, lineHeight: '1.55556', padding: '24px', fontSize: '1.125rem' }}>
          危险操作
        </DialogTitle>
        <Divider />
        <DialogContent>您正在删除自己的帐户，将清空所有数据且不可恢复</DialogContent>
        <DialogActions>
          <Button onClick={() => setShowAccountDeleteModal(false)}>取消</Button>
          <Button
            sx={{ color: 'error.main' }}
            onClick={async () => {
              setShowAccountDeleteModal(false);
            }}
          >
            确定
          </Button>
        </DialogActions>
      </Dialog>
      <WechatModal open={openWechat} handleClose={handleWechatClose} wechatLogin={bindWeChat} qrCode={status.wechat_qrcode} />
      <EmailModal
        open={openEmail}
        turnstileToken={turnstileToken}
        handleClose={() => {
          setOpenEmail(false);
        }}
      />
    </>
  );
}
