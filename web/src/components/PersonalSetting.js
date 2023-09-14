import React, { useContext, useEffect, useState } from 'react';
import { Button, Divider, Form, Header, Image, Message, Modal } from 'semantic-ui-react';
import { Link, useNavigate } from 'react-router-dom';
import { API, copy, showError, showInfo, showNotice, showSuccess } from '../helpers';
import Turnstile from 'react-turnstile';
import { UserContext } from '../context/User';
import { onGitHubOAuthClicked } from './utils';

const PersonalSetting = () => {
  const [userState, userDispatch] = useContext(UserContext);
  let navigate = useNavigate();

  const [inputs, setInputs] = useState({
    wechat_verification_code: '',
    email_verification_code: '',
    email: '',
    self_account_deletion_confirmation: ''
  });
  const [status, setStatus] = useState({});
  const [showWeChatBindModal, setShowWeChatBindModal] = useState(false);
  const [showEmailBindModal, setShowEmailBindModal] = useState(false);
  const [showAccountDeleteModal, setShowAccountDeleteModal] = useState(false);
  const [turnstileEnabled, setTurnstileEnabled] = useState(false);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [turnstileToken, setTurnstileToken] = useState('');
  const [loading, setLoading] = useState(false);
  const [disableButton, setDisableButton] = useState(false);
  const [countdown, setCountdown] = useState(30);
  const [affLink, setAffLink] = useState("");
  const [systemToken, setSystemToken] = useState("");

  useEffect(() => {
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      setStatus(status);
      if (status.turnstile_check) {
        setTurnstileEnabled(true);
        setTurnstileSiteKey(status.turnstile_site_key);
      }
    }
  }, []);

  useEffect(() => {
    let countdownInterval = null;
    if (disableButton && countdown > 0) {
      countdownInterval = setInterval(() => {
        setCountdown(countdown - 1);
      }, 1000);
    } else if (countdown === 0) {
      setDisableButton(false);
      setCountdown(30);
    }
    return () => clearInterval(countdownInterval); // Clean up on unmount
  }, [disableButton, countdown]);

  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const generateAccessToken = async () => {
    const res = await API.get('/api/user/token');
    const { success, message, data } = res.data;
    if (success) {
      setSystemToken(data);
      setAffLink(""); 
      await copy(data);
      showSuccess(`令牌已重置并已复制到剪贴板`);
    } else {
      showError(message);
    }
  };

  const getAffLink = async () => {
    const res = await API.get('/api/user/aff');
    const { success, message, data } = res.data;
    if (success) {
      let link = `${window.location.origin}/register?aff=${data}`;
      setAffLink(link);
      setSystemToken("");
      await copy(link);
      showSuccess(`邀请链接已复制到剪切板`);
    } else {
      showError(message);
    }
  };

  const handleAffLinkClick = async (e) => {
    e.target.select();
    await copy(e.target.value);
    showSuccess(`邀请链接已复制到剪切板`);
  };

  const handleSystemTokenClick = async (e) => {
    e.target.select();
    await copy(e.target.value);
    showSuccess(`系统令牌已复制到剪切板`);
  };

  const deleteAccount = async () => {
    if (inputs.self_account_deletion_confirmation !== userState.user.username) {
      showError('请输入你的账户名以确认删除！');
      return;
    }

    const res = await API.delete('/api/user/self');
    const { success, message } = res.data;

    if (success) {
      showSuccess('账户已删除！');
      await API.get('/api/user/logout');
      userDispatch({ type: 'logout' });
      localStorage.removeItem('user');
      navigate('/login');
    } else {
      showError(message);
    }
  };

  const bindWeChat = async () => {
    if (inputs.wechat_verification_code === '') return;
    const res = await API.get(
      `/api/oauth/wechat/bind?code=${inputs.wechat_verification_code}`
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess('微信账户绑定成功！');
      setShowWeChatBindModal(false);
    } else {
      showError(message);
    }
  };

  const sendVerificationCode = async () => {
    setDisableButton(true);
    if (inputs.email === '') return;
    if (turnstileEnabled && turnstileToken === '') {
      showInfo('请稍后几秒重试，Turnstile 正在检查用户环境！');
      return;
    }
    setLoading(true);
    const res = await API.get(
      `/api/verification?email=${inputs.email}&turnstile=${turnstileToken}`
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess('验证码发送成功，请检查邮箱！');
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const bindEmail = async () => {
    if (inputs.email_verification_code === '') return;
    setLoading(true);
    const res = await API.get(
      `/api/oauth/email/bind?email=${inputs.email}&code=${inputs.email_verification_code}`
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess('邮箱账户绑定成功！');
      setShowEmailBindModal(false);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  return (
    <div style={{ lineHeight: '40px' }}>
      <Header as='h3'>通用设置</Header>
      <Message>
        注意，此处生成的令牌用于系统管理，而非用于请求 OpenAI 相关的服务，请知悉。
      </Message>
      <Button as={Link} to={`/user/edit/`}>
        更新个人信息
      </Button>
      <Button onClick={generateAccessToken}>生成系统访问令牌</Button>
      <Button onClick={getAffLink}>复制邀请链接</Button>
      <Button onClick={() => {
        setShowAccountDeleteModal(true);
      }}>删除个人账户</Button>
      
      {systemToken && (
        <Form.Input 
          fluid 
          readOnly 
          value={systemToken} 
          onClick={handleSystemTokenClick}
          style={{ marginTop: '10px' }}
        />
      )}
      {affLink && (
        <Form.Input 
          fluid 
          readOnly 
          value={affLink} 
          onClick={handleAffLinkClick}
          style={{ marginTop: '10px' }}
        />
      )}
      <Divider />
      <Header as='h3'>账号绑定</Header>
      {
        status.wechat_login && (
          <Button
            onClick={() => {
              setShowWeChatBindModal(true);
            }}
          >
            绑定微信账号
          </Button>
        )
      }
      <Modal
        onClose={() => setShowWeChatBindModal(false)}
        onOpen={() => setShowWeChatBindModal(true)}
        open={showWeChatBindModal}
        size={'mini'}
      >
        <Modal.Content>
          <Modal.Description>
            <Image src={status.wechat_qrcode} fluid />
            <div style={{ textAlign: 'center' }}>
              <p>
                微信扫码关注公众号，输入「验证码」获取验证码（三分钟内有效）
              </p>
            </div>
            <Form size='large'>
              <Form.Input
                fluid
                placeholder='验证码'
                name='wechat_verification_code'
                value={inputs.wechat_verification_code}
                onChange={handleInputChange}
              />
              <Button color='' fluid size='large' onClick={bindWeChat}>
                绑定
              </Button>
            </Form>
          </Modal.Description>
        </Modal.Content>
      </Modal>
      {
        status.github_oauth && (
          <Button onClick={()=>{onGitHubOAuthClicked(status.github_client_id)}}>绑定 GitHub 账号</Button>
        )
      }
      <Button
        onClick={() => {
          setShowEmailBindModal(true);
        }}
      >
        绑定邮箱地址
      </Button>
      <Modal
        onClose={() => setShowEmailBindModal(false)}
        onOpen={() => setShowEmailBindModal(true)}
        open={showEmailBindModal}
        size={'tiny'}
        style={{ maxWidth: '450px' }}
      >
        <Modal.Header>绑定邮箱地址</Modal.Header>
        <Modal.Content>
          <Modal.Description>
            <Form size='large'>
              <Form.Input
                fluid
                placeholder='输入邮箱地址'
                onChange={handleInputChange}
                name='email'
                type='email'
                action={
                  <Button onClick={sendVerificationCode} disabled={disableButton || loading}>
                    {disableButton ? `重新发送(${countdown})` : '获取验证码'}
                  </Button>
                }
              />
              <Form.Input
                fluid
                placeholder='验证码'
                name='email_verification_code'
                value={inputs.email_verification_code}
                onChange={handleInputChange}
              />
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
              <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '1rem' }}>
              <Button
                color=''
                fluid
                size='large'
                onClick={bindEmail}
                loading={loading}
              >
                确认绑定
              </Button>
              <div style={{ width: '1rem' }}></div> 
              <Button
                fluid
                size='large'
                onClick={() => setShowEmailBindModal(false)}
              >
                取消
              </Button>
              </div>
            </Form>
          </Modal.Description>
        </Modal.Content>
      </Modal>
      <Modal
        onClose={() => setShowAccountDeleteModal(false)}
        onOpen={() => setShowAccountDeleteModal(true)}
        open={showAccountDeleteModal}
        size={'tiny'}
        style={{ maxWidth: '450px' }}
      >
        <Modal.Header>危险操作</Modal.Header>
        <Modal.Content>
        <Message>您正在删除自己的帐户，将清空所有数据且不可恢复</Message>
          <Modal.Description>
            <Form size='large'>
              <Form.Input
                fluid
                placeholder={`输入你的账户名 ${userState?.user?.username} 以确认删除`}
                name='self_account_deletion_confirmation'
                value={inputs.self_account_deletion_confirmation}
                onChange={handleInputChange}
              />
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
              <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '1rem' }}>
                <Button
                  color='red'
                  fluid
                  size='large'
                  onClick={deleteAccount}
                  loading={loading}
                >
                  确认删除
                </Button>
                <div style={{ width: '1rem' }}></div>
                <Button
                  fluid
                  size='large'
                  onClick={() => setShowAccountDeleteModal(false)}
                >
                  取消
                </Button>
              </div>
            </Form>
          </Modal.Description>
        </Modal.Content>
      </Modal>
    </div>
  );
};

export default PersonalSetting;
