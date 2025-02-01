import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Divider, Form, Grid, Header, Modal, Message } from 'semantic-ui-react';
import { API, removeTrailingSlash, showError } from '../helpers';

const SystemSetting = () => {
  const { t } = useTranslation();
  let [inputs, setInputs] = useState({
    PasswordLoginEnabled: '',
    PasswordRegisterEnabled: '',
    EmailVerificationEnabled: '',
    GitHubOAuthEnabled: '',
    GitHubClientId: '',
    GitHubClientSecret: '',
    LarkClientId: '',
    LarkClientSecret: '',
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
    MessagePusherAddress: '',
    MessagePusherToken: '',
    TurnstileCheckEnabled: '',
    TurnstileSiteKey: '',
    TurnstileSecretKey: '',
    RegisterEnabled: '',
    EmailDomainRestrictionEnabled: '',
    EmailDomainWhitelist: ''
  });
  const [originInputs, setOriginInputs] = useState({});
  let [loading, setLoading] = useState(false);
  const [EmailDomainWhitelist, setEmailDomainWhitelist] = useState([]);
  const [restrictedDomainInput, setRestrictedDomainInput] = useState('');
  const [showPasswordWarningModal, setShowPasswordWarningModal] = useState(false);

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        newInputs[item.key] = item.value;
      });
      setInputs({
        ...newInputs,
        EmailDomainWhitelist: newInputs.EmailDomainWhitelist.split(',')
      });
      setOriginInputs(newInputs);

      setEmailDomainWhitelist(newInputs.EmailDomainWhitelist.split(',').map((item) => {
        return { key: item, text: item, value: item };
      }));
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
      case 'WeChatAuthEnabled':
      case 'TurnstileCheckEnabled':
      case 'EmailDomainRestrictionEnabled':
      case 'RegisterEnabled':
        value = inputs[key] === 'true' ? 'false' : 'true';
        break;
      default:
        break;
    }
    const res = await API.put('/api/option/', {
      key,
      value
    });
    const { success, message } = res.data;
    if (success) {
      if (key === 'EmailDomainWhitelist') {
        value = value.split(',');
      }
      setInputs((inputs) => ({
        ...inputs, [key]: value
      }));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const handleInputChange = async (e, { name, value }) => {
    if (name === 'PasswordLoginEnabled' && inputs[name] === 'true') {
      // block disabling password login
      setShowPasswordWarningModal(true);
      return;
    }
    if (
      name === 'Notice' ||
      name.startsWith('SMTP') ||
      name === 'ServerAddress' ||
      name === 'GitHubClientId' ||
      name === 'GitHubClientSecret' ||
      name === 'LarkClientId' ||
      name === 'LarkClientSecret' ||
      name === 'WeChatServerAddress' ||
      name === 'WeChatServerToken' ||
      name === 'WeChatAccountQRCodeImageURL' ||
      name === 'TurnstileSiteKey' ||
      name === 'TurnstileSecretKey' ||
      name === 'EmailDomainWhitelist'
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


  const submitEmailDomainWhitelist = async () => {
    if (
      originInputs['EmailDomainWhitelist'] !== inputs.EmailDomainWhitelist.join(',') &&
      inputs.SMTPToken !== ''
    ) {
      await updateOption('EmailDomainWhitelist', inputs.EmailDomainWhitelist.join(','));
    }
  };

  const submitWeChat = async () => {
    if (originInputs['WeChatServerAddress'] !== inputs.WeChatServerAddress) {
      await updateOption(
        'WeChatServerAddress',
        removeTrailingSlash(inputs.WeChatServerAddress)
      );
    }
    if (
      originInputs['WeChatAccountQRCodeImageURL'] !==
      inputs.WeChatAccountQRCodeImageURL
    ) {
      await updateOption(
        'WeChatAccountQRCodeImageURL',
        inputs.WeChatAccountQRCodeImageURL
      );
    }
    if (
      originInputs['WeChatServerToken'] !== inputs.WeChatServerToken &&
      inputs.WeChatServerToken !== ''
    ) {
      await updateOption('WeChatServerToken', inputs.WeChatServerToken);
    }
  };

  const submitMessagePusher = async () => {
    if (originInputs['MessagePusherAddress'] !== inputs.MessagePusherAddress) {
      await updateOption(
        'MessagePusherAddress',
        removeTrailingSlash(inputs.MessagePusherAddress)
      );
    }
    if (
      originInputs['MessagePusherToken'] !== inputs.MessagePusherToken &&
      inputs.MessagePusherToken !== ''
    ) {
      await updateOption('MessagePusherToken', inputs.MessagePusherToken);
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

   const submitLarkOAuth = async () => {
    if (originInputs['LarkClientId'] !== inputs.LarkClientId) {
      await updateOption('LarkClientId', inputs.LarkClientId);
    }
    if (
      originInputs['LarkClientSecret'] !== inputs.LarkClientSecret &&
      inputs.LarkClientSecret !== ''
    ) {
      await updateOption('LarkClientSecret', inputs.LarkClientSecret);
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

  const submitNewRestrictedDomain = () => {
    const localDomainList = inputs.EmailDomainWhitelist;
    if (restrictedDomainInput !== '' && !localDomainList.includes(restrictedDomainInput)) {
      setRestrictedDomainInput('');
      setInputs({
        ...inputs,
        EmailDomainWhitelist: [...localDomainList, restrictedDomainInput],
      });
      setEmailDomainWhitelist([...EmailDomainWhitelist, {
        key: restrictedDomainInput,
        text: restrictedDomainInput,
        value: restrictedDomainInput,
      }]);
    }
  }

  return (
    <Grid columns={1}>
      <Grid.Column>
        <Form loading={loading}>
          <Header as='h3'>{t('setting.system.general.title')}</Header>
          <Form.Group widths='equal'>
            <Form.Input
              label={t('setting.system.general.server_address')}
              placeholder={t('setting.system.general.server_address_placeholder')}
              value={inputs.ServerAddress}
              name='ServerAddress'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Button onClick={submitServerAddress}>
            {t('setting.system.general.buttons.update')}
          </Form.Button>
          <Divider />
          <Header as='h3'>{t('setting.system.login.title')}</Header>
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.PasswordLoginEnabled === 'true'}
              label={t('setting.system.login.password_login')}
              name='PasswordLoginEnabled'
              onChange={handleInputChange}
            />
            {showPasswordWarningModal && (
              <Modal
                open={showPasswordWarningModal}
                onClose={() => setShowPasswordWarningModal(false)}
                size={'tiny'}
                style={{ maxWidth: '450px' }}
              >
                <Modal.Header>{t('setting.system.password_login.warning.title')}</Modal.Header>
                <Modal.Content>
                  <p>{t('setting.system.password_login.warning.content')}</p>
                </Modal.Content>
                <Modal.Actions>
                  <Button onClick={() => setShowPasswordWarningModal(false)}>
                    {t('setting.system.password_login.warning.buttons.cancel')}
                  </Button>
                  <Button
                    color='yellow'
                    onClick={async () => {
                      setShowPasswordWarningModal(false);
                      await updateOption('PasswordLoginEnabled', 'false');
                    }}
                  >
                    {t('setting.system.password_login.warning.buttons.confirm')}
                  </Button>
                </Modal.Actions>
              </Modal>
            )}
            <Form.Checkbox
              checked={inputs.PasswordRegisterEnabled === 'true'}
              label={t('setting.system.login.password_register')}
              name='PasswordRegisterEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.EmailVerificationEnabled === 'true'}
              label={t('setting.system.login.email_verification')}
              name='EmailVerificationEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.GitHubOAuthEnabled === 'true'}
              label={t('setting.system.login.github_oauth')}
              name='GitHubOAuthEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.WeChatAuthEnabled === 'true'}
              label={t('setting.system.login.wechat_login')}
              name='WeChatAuthEnabled'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.RegisterEnabled === 'true'}
              label={t('setting.system.login.registration')}
              name='RegisterEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.TurnstileCheckEnabled === 'true'}
              label={t('setting.system.login.turnstile')}
              name='TurnstileCheckEnabled'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Divider />
          <Header as='h3'>{t('setting.system.email_restriction.title')}</Header>
          <Message>{t('setting.system.email_restriction.subtitle')}</Message>
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.EmailDomainRestrictionEnabled === 'true'}
              label={t('setting.system.email_restriction.enable')}
              name='EmailDomainRestrictionEnabled'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Group widths={3}>
            <Form.Input
              label={t('setting.system.email_restriction.add_domain')}
              placeholder={t('setting.system.email_restriction.add_domain_placeholder')}
              value={restrictedDomainInput}
              onChange={(e, { value }) => {
                setRestrictedDomainInput(value);
              }}
              action={
                <Button onClick={() => {
                  if (restrictedDomainInput === '') return;
                  setEmailDomainWhitelist([...EmailDomainWhitelist, {
                    key: restrictedDomainInput,
                    text: restrictedDomainInput,
                    value: restrictedDomainInput
                  }]);
                  setRestrictedDomainInput('');
                }}>
                  {t('setting.system.email_restriction.buttons.fill')}
                </Button>
              }
            />
          </Form.Group>
          <Form.Dropdown
            label={t('setting.system.email_restriction.allowed_domains')}
            placeholder={t('setting.system.email_restriction.allowed_domains')}
            fluid
            multiple
            search
            selection
            allowAdditions
            value={EmailDomainWhitelist.map(item => item.value)}
            options={EmailDomainWhitelist}
            onAddItem={(e, { value }) => {
              setEmailDomainWhitelist([...EmailDomainWhitelist, {
                key: value,
                text: value,
                value: value
              }]);
            }}
            onChange={(e, { value }) => {
              let newEmailDomainWhitelist = [];
              value.forEach((item) => {
                newEmailDomainWhitelist.push({
                  key: item,
                  text: item,
                  value: item
                });
              });
              setEmailDomainWhitelist(newEmailDomainWhitelist);
            }}
          />
          <Form.Button onClick={submitEmailDomainWhitelist}>
            {t('setting.system.email_restriction.buttons.save')}
          </Form.Button>

          <Divider />
          <Header as='h3'>{t('setting.system.smtp.title')}</Header>
          <Message>{t('setting.system.smtp.subtitle')}</Message>
          <Form.Group widths={3}>
            <Form.Input
              label={t('setting.system.smtp.server')}
              placeholder={t('setting.system.smtp.server_placeholder')}
              name='SMTPServer'
              onChange={handleInputChange}
              value={inputs.SMTPServer}
            />
            <Form.Input
              label={t('setting.system.smtp.port')}
              placeholder={t('setting.system.smtp.port_placeholder')}
              name='SMTPPort'
              onChange={handleInputChange}
              value={inputs.SMTPPort}
            />
            <Form.Input
              label={t('setting.system.smtp.account')}
              placeholder={t('setting.system.smtp.account_placeholder')}
              name='SMTPAccount'
              onChange={handleInputChange}
              value={inputs.SMTPAccount}
            />
          </Form.Group>
          <Form.Group widths={3}>
            <Form.Input
              label={t('setting.system.smtp.from')}
              placeholder={t('setting.system.smtp.from_placeholder')}
              name='SMTPFrom'
              onChange={handleInputChange}
              value={inputs.SMTPFrom}
            />
            <Form.Input
              label={t('setting.system.smtp.token')}
              placeholder={t('setting.system.smtp.token_placeholder')}
              name='SMTPToken'
              onChange={handleInputChange}
              type='password'
              value={inputs.SMTPToken}
            />
          </Form.Group>
          <Form.Button onClick={submitSMTP}>
            {t('setting.system.smtp.buttons.save')}
          </Form.Button>

          <Divider />
          <Header as='h3'>{t('setting.system.github.title')}</Header>
          <Message>
            {t('setting.system.github.subtitle')}
            <a href='https://github.com/settings/developers' target='_blank'>
              {t('setting.system.github.manage_link')}
            </a>
            {t('setting.system.github.manage_text')}
          </Message>
          <Message>
            {t('setting.system.github.url_notice', {
              server_url: originInputs.ServerAddress,
              callback_url: `${originInputs.ServerAddress}/oauth/github`
            })}
          </Message>
          <Form.Group widths={3}>
            <Form.Input
              label={t('setting.system.github.client_id')}
              placeholder={t('setting.system.github.client_id_placeholder')}
              name='GitHubClientId'
              onChange={handleInputChange}
              value={inputs.GitHubClientId}
            />
            <Form.Input
              label={t('setting.system.github.client_secret')}
              placeholder={t('setting.system.github.client_secret_placeholder')}
              name='GitHubClientSecret'
              onChange={handleInputChange}
              type='password'
              value={inputs.GitHubClientSecret}
            />
          </Form.Group>
          <Form.Button onClick={submitGitHubOAuth}>
            {t('setting.system.github.buttons.save')}
          </Form.Button>

          <Divider />
          <Header as='h3'>
            {t('setting.system.lark.title')}
            <Header.Subheader>
              {t('setting.system.lark.subtitle')}
              <a href='https://open.feishu.cn/app' target='_blank'>
                {t('setting.system.lark.manage_link')}
              </a>
              {t('setting.system.lark.manage_text')}
            </Header.Subheader>
          </Header>
          <Message>
            {t('setting.system.lark.url_notice', {
              server_url: inputs.ServerAddress,
              callback_url: `${inputs.ServerAddress}/oauth/lark`
            })}
          </Message>
          <Form.Group widths={3}>
            <Form.Input
              label={t('setting.system.lark.client_id')}
              name='LarkClientId'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.LarkClientId}
              placeholder={t('setting.system.lark.client_id_placeholder')}
            />
            <Form.Input
              label={t('setting.system.lark.client_secret')}
              name='LarkClientSecret'
              onChange={handleInputChange}
              type='password'
              autoComplete='new-password'
              value={inputs.LarkClientSecret}
              placeholder={t('setting.system.lark.client_secret_placeholder')}
            />
          </Form.Group>
          <Form.Button onClick={submitLarkOAuth}>
            {t('setting.system.lark.buttons.save')}
          </Form.Button>

          <Divider />
          <Header as='h3'>
            {t('setting.system.wechat.title')}
            <Header.Subheader>
              {t('setting.system.wechat.subtitle')}
              <a
                href='https://github.com/songquanpeng/wechat-server'
                target='_blank'
              >
                {t('setting.system.wechat.learn_more')}
              </a>
            </Header.Subheader>
          </Header>
          <Form.Group widths={3}>
            <Form.Input
              label={t('setting.system.wechat.server_address')}
              name='WeChatServerAddress'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.WeChatServerAddress}
              placeholder={t('setting.system.wechat.server_address_placeholder')}
            />
            <Form.Input
              label={t('setting.system.wechat.token')}
              name='WeChatServerToken'
              onChange={handleInputChange}
              type='password'
              autoComplete='new-password'
              value={inputs.WeChatServerToken}
              placeholder={t('setting.system.wechat.token_placeholder')}
            />
            <Form.Input
              label={t('setting.system.wechat.qrcode')}
              name='WeChatAccountQRCodeImageURL'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.WeChatAccountQRCodeImageURL}
              placeholder={t('setting.system.wechat.qrcode_placeholder')}
            />
          </Form.Group>
          <Form.Button onClick={submitWeChat}>
            {t('setting.system.wechat.buttons.save')}
          </Form.Button>

          <Divider />
          <Header as='h3'>
            {t('setting.system.turnstile.title')}
            <Header.Subheader>
              {t('setting.system.turnstile.subtitle')}
              <a href='https://dash.cloudflare.com/' target='_blank'>
                {t('setting.system.turnstile.manage_link')}
              </a>
              {t('setting.system.turnstile.manage_text')}
            </Header.Subheader>
          </Header>
          <Form.Group widths={3}>
            <Form.Input
              label={t('setting.system.turnstile.site_key')}
              name='TurnstileSiteKey'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.TurnstileSiteKey}
              placeholder={t('setting.system.turnstile.site_key_placeholder')}
            />
            <Form.Input
              label={t('setting.system.turnstile.secret_key')}
              name='TurnstileSecretKey'
              onChange={handleInputChange}
              type='password'
              autoComplete='new-password'
              value={inputs.TurnstileSecretKey}
              placeholder={t('setting.system.turnstile.secret_key_placeholder')}
            />
          </Form.Group>
          <Form.Button onClick={submitTurnstile}>
            {t('setting.system.turnstile.buttons.save')}
          </Form.Button>
        </Form>
      </Grid.Column>
    </Grid>
  );
};

export default SystemSetting;
