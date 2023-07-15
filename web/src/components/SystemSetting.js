import React, { useEffect, useState } from 'react';
import { Divider, Form, Grid, Header, Message } from 'semantic-ui-react';
import { API, removeTrailingSlash, showError, verifyJSON } from '../helpers';

const SystemSetting = () => {
  let [inputs, setInputs] = useState({
    PasswordLoginEnabled: '',
    PasswordRegisterEnabled: '',
    EmailVerificationEnabled: '',
    GitHubOAuthEnabled: '',
    DiscordOAuthEnabled: '',
    GitHubClientId: '',
    GitHubClientSecret: '',
    DiscordClientId: '',
    DiscordClientSecret: '',
    Notice: '',
    SMTPServer: '',
    SMTPPort: '',
    SMTPAccount: '',
    SMTPFrom: '',
    SMTPToken: '',
    ServerAddress: '',
    Footer: '',
    WeChatAuthEnabled: '',
    WeChatServerAddress: '',
    WeChatServerToken: '',
    WeChatAccountQRCodeImageURL: '',
    TurnstileCheckEnabled: '',
    TurnstileSiteKey: '',
    TurnstileSecretKey: '',
    RegisterEnabled: '',
  });
  const [originInputs, setOriginInputs] = useState({});
  let [loading, setLoading] = useState(false);

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        newInputs[item.key] = item.value;
      });
      setInputs(newInputs);
      setOriginInputs(newInputs);
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    getOptions().then();
  }, []);

  const updateOption = async (key, value) => {
    setLoading(true);
    switch (key) {
      case 'PasswordLoginEnabled':
      case 'PasswordRegisterEnabled':
      case 'EmailVerificationEnabled':
      case 'GitHubOAuthEnabled':
      case 'DiscordOAuthEnabled':
      case 'WeChatAuthEnabled':
      case 'TurnstileCheckEnabled':
      case 'RegisterEnabled':
        value = inputs[key] === 'true' ? 'false' : 'true';
        break;
      default:
        break;
    }
    const res = await API.put('/api/option/', {
      key,
      value,
    });
    const { success, message } = res.data;
    if (success) {
      setInputs((inputs) => ({ ...inputs, [key]: value }));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const handleInputChange = async (e, { name, value }) => {
    if (
      name === 'Notice' ||
      name.startsWith('SMTP') ||
      name === 'ServerAddress' ||
      name === 'DiscordClientId' ||
      name === 'DiscordClientSecret' ||
      name === 'GitHubClientId' ||
      name === 'GitHubClientSecret' ||
      name === 'WeChatServerAddress' ||
      name === 'WeChatServerToken' ||
      name === 'WeChatAccountQRCodeImageURL' ||
      name === 'TurnstileSiteKey' ||
      name === 'TurnstileSecretKey'
    ) {
      setInputs((inputs) => ({ ...inputs, [name]: value }));
    } else {
      await updateOption(name, value);
    }
  };

  const submitServerAddress = async () => {
    let ServerAddress = removeTrailingSlash(inputs.ServerAddress);
    await updateOption('ServerAddress', ServerAddress);
  };

  const submitSMTP = async () => {
    if (originInputs['SMTPServer'] !== inputs.SMTPServer) {
      await updateOption('SMTPServer', inputs.SMTPServer);
    }
    if (originInputs['SMTPAccount'] !== inputs.SMTPAccount) {
      await updateOption('SMTPAccount', inputs.SMTPAccount);
    }
    if (originInputs['SMTPFrom'] !== inputs.SMTPFrom) {
      await updateOption('SMTPFrom', inputs.SMTPFrom);
    }
    if (
      originInputs['SMTPPort'] !== inputs.SMTPPort &&
      inputs.SMTPPort !== ''
    ) {
      await updateOption('SMTPPort', inputs.SMTPPort);
    }
    if (
      originInputs['SMTPToken'] !== inputs.SMTPToken &&
      inputs.SMTPToken !== ''
    ) {
      await updateOption('SMTPToken', inputs.SMTPToken);
    }
  };

  const submitWeChat = async () => {
    if (originInputs['WeChatServerAddress'] !== inputs.WeChatServerAddress) {
      await updateOption(
        'WeChatServerAddress',
        removeTrailingSlash(inputs.WeChatServerAddress),
      );
    }
    if (
      originInputs['WeChatAccountQRCodeImageURL'] !==
      inputs.WeChatAccountQRCodeImageURL
    ) {
      await updateOption(
        'WeChatAccountQRCodeImageURL',
        inputs.WeChatAccountQRCodeImageURL,
      );
    }
    if (
      originInputs['WeChatServerToken'] !== inputs.WeChatServerToken &&
      inputs.WeChatServerToken !== ''
    ) {
      await updateOption('WeChatServerToken', inputs.WeChatServerToken);
    }
  };

  const submitGitHubOAuth = async () => {
    if (originInputs['GitHubClientId'] !== inputs.GitHubClientId) {
      await updateOption('GitHubClientId', inputs.GitHubClientId);
    }
    if (
      originInputs['GitHubClientSecret'] !== inputs.GitHubClientSecret &&
      inputs.GitHubClientSecret !== ''
    ) {
      await updateOption('GitHubClientSecret', inputs.GitHubClientSecret);
    }
  };

  const submitDiscordOAuth = async () => {
    if (originInputs['DiscordClientId'] !== inputs.DiscordClientId) {
      await updateOption('DiscordClientId', inputs.DiscordClientId);
    }
    if (
      originInputs['DiscordClientSecret'] !== inputs.DiscordClientSecret &&
      inputs.DiscordClientSecret !== ''
    ) {
      await updateOption('DiscordClientSecret', inputs.DiscordClientSecret);
    }
  };

  const submitTurnstile = async () => {
    if (originInputs['TurnstileSiteKey'] !== inputs.TurnstileSiteKey) {
      await updateOption('TurnstileSiteKey', inputs.TurnstileSiteKey);
    }
    if (
      originInputs['TurnstileSecretKey'] !== inputs.TurnstileSecretKey &&
      inputs.TurnstileSecretKey !== ''
    ) {
      await updateOption('TurnstileSecretKey', inputs.TurnstileSecretKey);
    }
  };

  return (
    <Grid columns={1}>
      <Grid.Column>
        <Form loading={loading}>
          <Header as='h3'>General Settings</Header>
          <Form.Group widths='equal'>
            <Form.Input
              label='Server Address'
              placeholder='For example：https://yourdomain.com'
              value={inputs.ServerAddress}
              name='ServerAddress'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Button onClick={submitServerAddress}>
            Update Server Address
          </Form.Button>
          <Divider />
          <Header as='h3'>Configure Login/Registration</Header>
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.PasswordLoginEnabled === 'true'}
              label='Allow login via password'
              name='PasswordLoginEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.PasswordRegisterEnabled === 'true'}
              label='Allow registration via password'
              name='PasswordRegisterEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.EmailVerificationEnabled === 'true'}
              label='Email verification is required when registering via password'
              name='EmailVerificationEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.GitHubOAuthEnabled === 'true'}
              label='Allow login & registration via GitHub account'
              name='GitHubOAuthEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.DiscordOAuthEnabled === 'true'}
              label='允许通过 Discord 账户登录和注册'
              name='DiscordOAuthEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.WeChatAuthEnabled === 'true'}
              label='Allow login & registration via WeChat'
              name='WeChatAuthEnabled'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.RegisterEnabled === 'true'}
              label='Allow new user registration (if this option is off, new users will not be able to register in any way）'
              name='RegisterEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.TurnstileCheckEnabled === 'true'}
              label='Enable Turnstile user verification'
              name='TurnstileCheckEnabled'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Divider />
          <Header as='h3'>
            Configure SMTP
            <Header.Subheader>
              To support the system email sending
            </Header.Subheader>
          </Header>
          <Form.Group widths={3}>
            <Form.Input
              label='SMTP Server Address'
              name='SMTPServer'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.SMTPServer}
              placeholder='For example: smtp.qq.com'
            />
            <Form.Input
              label='SMTP Port'
              name='SMTPPort'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.SMTPPort}
              placeholder='Default: 587'
            />
            <Form.Input
              label='SMTP Account'
              name='SMTPAccount'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.SMTPAccount}
              placeholder='Usually an email address'
            />
          </Form.Group>
          <Form.Group widths={3}>
            <Form.Input
              label='SMTP Sender email'
              name='SMTPFrom'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.SMTPFrom}
              placeholder='Usually consistent with the email address'
            />
            <Form.Input
              label='SMTP Access Credential'
              name='SMTPToken'
              onChange={handleInputChange}
              type='password'
              autoComplete='new-password'
              value={inputs.SMTPToken}
              placeholder='Sensitive information will not be displayed in the frontend'
            />
          </Form.Group>
          <Form.Button onClick={submitSMTP}>Save SMTP Settings</Form.Button>
          <Divider />
          <Header as='h3'>
            Configure Discord OAuth App
            <Header.Subheader>
              To support login & registration via GitHub，
              <a
                href='https://discord.com/developers/applications'
                target='_blank'
              >
                Click here
              </a>
              Manage your Discord OAuth App
            </Header.Subheader>
          </Header>
          <Message>
            Fill in the Homepage URL <code>{inputs.ServerAddress}</code>
            ，Fill in the Authorization callback URL{' '}
            <code>{`${inputs.ServerAddress}/oauth/discord`}</code>
          </Message>
          <Form.Group widths={3}>
            <Form.Input
              label='Discord Client ID'
              name='DiscordClientId'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.DiscordClientId}
              placeholder='Enter the ID of your registered Discord OAuth APP'
            />
            <Form.Input
              label='Discord Client Secret'
              name='DiscordClientSecret'
              onChange={handleInputChange}
              type='password'
              autoComplete='new-password'
              value={inputs.DiscordClientSecret}
              placeholder='Sensitive information will not be displayed in the frontend'
            />
          </Form.Group>
          <Form.Button onClick={submitDiscordOAuth}>
            Save Discord OAuth Settings
          </Form.Button>
          <Divider />
          <Header as='h3'>
            Configure GitHub OAuth App
            <Header.Subheader>
              To support login & registration via GitHub，
              <a href='https://github.com/settings/developers' target='_blank'>
                Click here
              </a>
              Manage your GitHub OAuth App
            </Header.Subheader>
          </Header>
          <Message>
            Fill in the Homepage URL <code>{inputs.ServerAddress}</code>
            ，Fill in the Authorization callback URL{' '}
            <code>{`${inputs.ServerAddress}/oauth/github`}</code>
          </Message>
          <Form.Group widths={3}>
            <Form.Input
              label='GitHub Client ID'
              name='GitHubClientId'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.GitHubClientId}
              placeholder='Enter your registered GitHub OAuth APP ID'
            />
            <Form.Input
              label='GitHub Client Secret'
              name='GitHubClientSecret'
              onChange={handleInputChange}
              type='password'
              autoComplete='new-password'
              value={inputs.GitHubClientSecret}
              placeholder='Sensitive information will not be displayed in the frontend'
            />
          </Form.Group>
          <Form.Button onClick={submitGitHubOAuth}>
            Save GitHub OAuth Settings
          </Form.Button>
          <Divider />
          <Header as='h3'>
            Configure WeChat Server
            <Header.Subheader>
              To support login & registration via WeChat，
              <a
                href='https://github.com/songquanpeng/wechat-server'
                target='_blank'
              >
                Click here
              </a>
              Learn about WeChat Server
            </Header.Subheader>
          </Header>
          <Form.Group widths={3}>
            <Form.Input
              label='WeChat Server Server Address'
              name='WeChatServerAddress'
              placeholder='For example：https://yourdomain.com'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.WeChatServerAddress}
            />
            <Form.Input
              label='WeChat Server Access Credential'
              name='WeChatServerToken'
              type='password'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.WeChatServerToken}
              placeholder='Sensitive information will not be displayed in the frontend'
            />
            <Form.Input
              label='WeChat Public Account QR Code Image Link'
              name='WeChatAccountQRCodeImageURL'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.WeChatAccountQRCodeImageURL}
              placeholder='Enter an image link'
            />
          </Form.Group>
          <Form.Button onClick={submitWeChat}>
            Save WeChat Server Settings
          </Form.Button>
          <Divider />
          <Header as='h3'>
            Configure Turnstile
            <Header.Subheader>
              To support user verification，
              <a href='https://dash.cloudflare.com/' target='_blank'>
                Click here
              </a>
              Manage your Turnstile Sites, recommend selecting Invisible Widget
              Type
            </Header.Subheader>
          </Header>
          <Form.Group widths={3}>
            <Form.Input
              label='Turnstile Site Key'
              name='TurnstileSiteKey'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.TurnstileSiteKey}
              placeholder='Enter your registered Turnstile Site Key'
            />
            <Form.Input
              label='Turnstile Secret Key'
              name='TurnstileSecretKey'
              onChange={handleInputChange}
              type='password'
              autoComplete='new-password'
              value={inputs.TurnstileSecretKey}
              placeholder='Sensitive information will not be displayed in the frontend'
            />
          </Form.Group>
          <Form.Button onClick={submitTurnstile}>
            Save Turnstile Settings
          </Form.Button>
        </Form>
      </Grid.Column>
    </Grid>
  );
};

export default SystemSetting;
