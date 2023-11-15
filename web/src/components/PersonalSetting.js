import React, { useContext, useEffect, useState } from 'react';
import { Button, Divider, Form, Header, Image, Message, Modal, Label } from 'semantic-ui-react';
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
  const [turnstileEnabled, setTurnstileEnabled] = useState(false);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [turnstileToken, setTurnstileToken] = useState('');
  const [loading, setLoading] = useState(false);
  const [disableButton, setDisableButton] = useState(false);
  const [countdown, setCountdown] = useState(30);
  const [affLink, setAffLink] = useState("");
  const [systemToken, setSystemToken] = useState("");
  const [userGroup, setUserGroup] = useState(null);
  const [quota, setQuota] = useState(null);
  const [usedQuota, setUsedQuota] = useState(null);
  const [requestCount, setRequestCount] = useState(null);
  const [githubID, setGithubID] = useState(null);
  const [username, setUsername] = useState(null);
  const [display_name, setDisplay_name] = useState(null);
  const [email, setEmail] = useState(null);

  const [modelsByOwner, setModelsByOwner] = useState({});
  const [key, setKey] = useState("");



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

  // Get user group
  useEffect(() => {
    (async () => {
      const res = await API.get(`/api/user/self`);
      if (res.data.success) {
        setDisplay_name(res.data.data.display_name);
        setUsername(res.data.data.username);
        setUserGroup(res.data.data.group);
        setEmail(res.data.data.email);
        setQuota(res.data.data.quota);
        setUsedQuota(res.data.data.used_quota);
        setRequestCount(res.data.data.request_count);
        setGithubID(res.data.data.github_id);
      } else {
        // Handle the error here
      }
    })();
  }, []);
  const quotaPerUnit = parseInt(localStorage.getItem("quota_per_unit"));

  // useEffect(() => {
  //   // 获取用户的第一个key
  //   const fetchFirstKey = async () => {
  //     try {
  //       const tokenRes = await API.get('/api/token/?p=0');
  //       if (tokenRes.data.success && tokenRes.data.data.length > 0) {
  //         const firstKey = tokenRes.data.data[0].key;
  //         setKey(firstKey);
  //         fetchModels(firstKey);
  //       } else {
  //         // 如果没有获取到key，显示提示消息
  //         showError('请先创建一个key');
  //       }
  //     } catch (error) {
  //       showError('获取key失败');
  //     }
  //   };

  //   // 获取模型信息
  //   const fetchModels = async (key) => {
  //     try {
  //       const modelsRes = await API.get('/v1/models', {
  //         headers: { Authorization: `Bearer sk-${key}` },
  //       });
  //       if (modelsRes.data && modelsRes.data.data) {
  //         const models = modelsRes.data.data;
  //         const groupedByOwner = models.reduce((acc, model) => {
  //           const owner = model.owned_by.toUpperCase();
  //           if (!acc[owner]) {
  //             acc[owner] = [];
  //           }
  //           acc[owner].push(model.id);
  //           return acc;
  //         }, {});

  //         // 对owners进行排序
  //         const sortedOwners = Object.keys(groupedByOwner).sort();
  //         const sortedGroupedByOwner = {};
  //         sortedOwners.forEach(owner => {
  //           // 对每个owner的models进行排序
  //           sortedGroupedByOwner[owner] = groupedByOwner[owner].sort();
  //         });
  //         setModelsByOwner(sortedGroupedByOwner);
  //       }
  //     } catch (error) {
  //       showError('获取模型失败');
  //     }
  //   };

  //   fetchFirstKey();
  // }, []);

  // // 定义一个固定宽度的label样式
  // const fixedWidthLabelStyle = {
  //   display: 'inline-block',
  //   minWidth: '150px', // 根据需要调整宽度
  //   textAlign: 'center',
  //   margin: '5px',
  // };


  const transformUserGroup = (group) => {
    switch (group) {
      case 'default':
        return '默认用户';
      case 'vip':
        return 'VIP用户';
      case 'svip':
        return 'SVIP用户';
      default:
        return group;
    }
  };
  const getUserGroupColor = (group) => {
    switch (group) {
      case 'default':
        return 'var(--czl-grayA)';
      case 'vip':
        return 'var(--czl-success-color)';
      case 'svip':
        return 'var(--czl-error-color)';
      default:
        return '';
    }
  };


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
    showSuccess(`系统Key已复制到剪切板`);
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
      <Header as='h3'>使用信息</Header>
<div style={{ marginBottom: '1em' }}>
  {display_name && (
    <Label basic style={{ margin: '0.5em', color: getUserGroupColor(userGroup), borderColor: getUserGroupColor(userGroup) }}>
      昵称：{transformUserGroup(display_name)}
    </Label>
  )}
  {username && (
    <Label basic style={{ margin: '0.5em', color: getUserGroupColor(userGroup), borderColor: getUserGroupColor(userGroup) }}>
      用户名：{transformUserGroup(username)}
    </Label>
  )}
  <Label basic style={{ margin: '0.5em', color: getUserGroupColor(userGroup), borderColor: getUserGroupColor(userGroup) }}>
    邮箱：{email ? email : "未绑定"}
  </Label>
  <Label basic style={{ margin: '0.5em', color: getUserGroupColor(userGroup), borderColor: getUserGroupColor(userGroup) }}>
    GitHub 账号：{githubID ? githubID : "未绑定"}
  </Label>
  <br></br>
  {userGroup && (
    <Label basic style={{ margin: '0.5em', color: getUserGroupColor(userGroup), borderColor: getUserGroupColor(userGroup) }}>
      用户组：{transformUserGroup(userGroup)}
    </Label>
  )}
  {quota !== null && quotaPerUnit && (
    <Label basic style={{ margin: '0.5em', color: getUserGroupColor(userGroup), borderColor: getUserGroupColor(userGroup) }}>
      额度：${(quota / quotaPerUnit).toFixed(2)}
    </Label>
  )}
  {usedQuota !== null && quotaPerUnit && (
    <Label basic style={{ margin: '0.5em', color: getUserGroupColor(userGroup), borderColor: getUserGroupColor(userGroup) }}>
      已用额度：${(usedQuota / quotaPerUnit).toFixed(2)}
    </Label>
  )}
  {requestCount !== null && (
    <Label basic style={{ margin: '0.5em', color: getUserGroupColor(userGroup), borderColor: getUserGroupColor(userGroup) }}>
      调用次数：{requestCount}
    </Label>
  )}
</div>


        <Divider />
      {/* <Header as='h3'>模型支持度</Header>
      {Object.keys(modelsByOwner).length > 0 ? (
        Object.entries(modelsByOwner).map(([owner, models], index) => (
          <div key={index}>
            <Header as='h4'>{owner}</Header>
            <div>
              {models.map((modelId, index) => (
                <Label key={index} style={fixedWidthLabelStyle}>
                  {modelId}
                </Label>
              ))}
            </div>
          </div>
        ))
      ) : (
        <Message info>
          <Message.Header>尚未绑定模型</Message.Header>
          <p>请先创建一个key</p>
        </Message>
      )}
      <Divider /> */}
      <Header as='h3'>通用设置</Header>
      {/* <Message>
        注意，此处生成的Key用于系统管理，而非用于请求 OpenAI 相关的服务，请知悉。
      </Message> */}
      <Button as={Link} to={`/user/edit/`}>
        更新个人信息
      </Button>
      {/* <Button onClick={generateAccessToken}>生成系统访问Key</Button> */}
      <Button onClick={getAffLink}>复制邀请链接</Button>

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
          <Button onClick={() => { onGitHubOAuthClicked(status.github_client_id) }}>绑定 GitHub 账号</Button>
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
    </div>
  );
};

export default PersonalSetting;
