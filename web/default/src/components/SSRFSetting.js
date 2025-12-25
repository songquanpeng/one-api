import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Divider,
  Form,
  Grid,
  Header,
  Message,
  Label,
  Radio,
} from 'semantic-ui-react';
import { API, showError, showSuccess } from '../helpers';

const SSRFSetting = () => {
  const { t } = useTranslation();
  let [inputs, setInputs] = useState({
    SSRFProtectionEnabled: '',
    SSRFAllowPrivateIP: '',
    SSRFApplyIPFilterForDomain: '',
    SSRFDomainListMode: 'blacklist',
    SSRFDomainList: '',
    SSRFIPListMode: 'blacklist',
    SSRFIPList: '',
    SSRFAllowedPorts: '',
  });
  const [originInputs, setOriginInputs] = useState({});
  let [loading, setLoading] = useState(false);

  // Domain list states
  const [domainInput, setDomainInput] = useState('');
  const [domainList, setDomainList] = useState([]);
  const [domainListMode, setDomainListMode] = useState('blacklist');

  // IP list states
  const [ipInput, setIPInput] = useState('');
  const [ipList, setIPList] = useState([]);
  const [ipListMode, setIPListMode] = useState('blacklist');

  // Port list states
  const [portInput, setPortInput] = useState('');
  const [allowedPorts, setAllowedPorts] = useState([]);

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

      // Parse domain list and mode
      if (newInputs.SSRFDomainListMode) {
        setDomainListMode(newInputs.SSRFDomainListMode);
      }
      if (newInputs.SSRFDomainList) {
        const list = newInputs.SSRFDomainList.split(',').filter(
          (d) => d.trim()
        );
        setDomainList(list);
      }

      // Parse IP list and mode
      if (newInputs.SSRFIPListMode) {
        setIPListMode(newInputs.SSRFIPListMode);
      }
      if (newInputs.SSRFIPList) {
        const list = newInputs.SSRFIPList.split(',').filter(
          (d) => d.trim()
        );
        setIPList(list);
      }

      // Parse allowed ports
      if (newInputs.SSRFAllowedPorts) {
        const ports = newInputs.SSRFAllowedPorts.split(',').filter(
          (p) => p.trim()
        );
        setAllowedPorts(ports);
      }
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    getOptions().then();
  }, []);

  const updateOption = async (key, value) => {
    setLoading(true);
    const res = await API.put('/api/option/', {
      key,
      value,
    });
    const { success, message } = res.data;
    if (success) {
      setInputs((inputs) => ({
        ...inputs,
        [key]: value,
      }));
      setOriginInputs((inputs) => ({
        ...inputs,
        [key]: value,
      }));
    } else {
      showError(message);
    }
    setLoading(false);
    return success;
  };

  const handleCheckboxChange = async (e, { name, checked }) => {
    const value = checked ? 'true' : 'false';
    await updateOption(name, value);
  };

  // Domain list handlers
  const addDomainToList = () => {
    if (domainInput.trim() && !domainList.includes(domainInput.trim())) {
      setDomainList([...domainList, domainInput.trim()]);
      setDomainInput('');
    }
  };

  const removeDomainFromList = (domain) => {
    setDomainList(domainList.filter((d) => d !== domain));
  };

  const handleDomainModeChange = (e, { value }) => {
    setDomainListMode(value);
  };

  const saveDomainSettings = async () => {
    const listValue = domainList.join(',');
    let success = true;

    // Save mode first
    if (originInputs.SSRFDomainListMode !== domainListMode) {
      success = (await updateOption('SSRFDomainListMode', domainListMode)) && success;
    }

    // Save list
    if (originInputs.SSRFDomainList !== listValue) {
      success = (await updateOption('SSRFDomainList', listValue)) && success;
    }

    if (success) {
      showSuccess(t('setting.ssrf.messages.save_success'));
    }
  };

  // IP list handlers
  const addIPToList = () => {
    if (ipInput.trim() && !ipList.includes(ipInput.trim())) {
      setIPList([...ipList, ipInput.trim()]);
      setIPInput('');
    }
  };

  const removeIPFromList = (ip) => {
    setIPList(ipList.filter((i) => i !== ip));
  };

  const handleIPModeChange = (e, { value }) => {
    setIPListMode(value);
  };

  const saveIPSettings = async () => {
    const listValue = ipList.join(',');
    let success = true;

    // Save mode first
    if (originInputs.SSRFIPListMode !== ipListMode) {
      success = (await updateOption('SSRFIPListMode', ipListMode)) && success;
    }

    // Save list
    if (originInputs.SSRFIPList !== listValue) {
      success = (await updateOption('SSRFIPList', listValue)) && success;
    }

    if (success) {
      showSuccess(t('setting.ssrf.messages.save_success'));
    }
  };

  // Port handlers
  const addPort = () => {
    if (portInput.trim() && !allowedPorts.includes(portInput.trim())) {
      setAllowedPorts([...allowedPorts, portInput.trim()]);
      setPortInput('');
    }
  };

  const removePort = (port) => {
    setAllowedPorts(allowedPorts.filter((p) => p !== port));
  };

  const saveAllowedPorts = async () => {
    const portsValue = allowedPorts.join(',');
    if (originInputs.SSRFAllowedPorts !== portsValue) {
      const success = await updateOption('SSRFAllowedPorts', portsValue);
      if (success) {
        showSuccess(t('setting.ssrf.messages.save_success'));
      }
    }
  };

  return (
    <Grid columns={1}>
      <Grid.Column>
        <Form loading={loading}>
          <Header as='h3'>{t('setting.ssrf.title')}</Header>
          <Message info>{t('setting.ssrf.description')}</Message>

          {/* Enable SSRF Protection */}
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.SSRFProtectionEnabled === 'true'}
              label={t('setting.ssrf.enable')}
              name='SSRFProtectionEnabled'
              onChange={handleCheckboxChange}
            />
          </Form.Group>
          <Message size='small'>
            {t('setting.ssrf.enable_description')}
          </Message>

          <Divider />

          {/* Allow Private IP */}
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.SSRFAllowPrivateIP === 'true'}
              label={t('setting.ssrf.allow_private_ip')}
              name='SSRFAllowPrivateIP'
              onChange={handleCheckboxChange}
            />
          </Form.Group>
          <Message size='small'>
            {t('setting.ssrf.allow_private_ip_description')}
          </Message>

          <Divider />

          {/* Apply IP Filter for Domain */}
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.SSRFApplyIPFilterForDomain === 'true'}
              label={t('setting.ssrf.apply_ip_filter_for_domain')}
              name='SSRFApplyIPFilterForDomain'
              onChange={handleCheckboxChange}
            />
          </Form.Group>
          <Message size='small'>
            {t('setting.ssrf.apply_ip_filter_for_domain_description')}
          </Message>

          <Divider />

          {/* Domain List */}
          <Header as='h4'>{t('setting.ssrf.domain_list.title')}</Header>
          <Message size='small'>
            {t('setting.ssrf.domain_list.description')}
          </Message>

          {/* Mode Toggle */}
          <Form.Group inline>
            <label>{t('setting.ssrf.mode_label')}</label>
            <Form.Field>
              <Radio
                label={t('setting.ssrf.mode.blacklist')}
                name='domainListMode'
                value='blacklist'
                checked={domainListMode === 'blacklist'}
                onChange={handleDomainModeChange}
              />
            </Form.Field>
            <Form.Field>
              <Radio
                label={t('setting.ssrf.mode.whitelist')}
                name='domainListMode'
                value='whitelist'
                checked={domainListMode === 'whitelist'}
                onChange={handleDomainModeChange}
              />
            </Form.Field>
          </Form.Group>

          <Message size='small' color={domainListMode === 'whitelist' ? 'yellow' : undefined}>
            {domainListMode === 'whitelist'
              ? t('setting.ssrf.mode.whitelist_description')
              : t('setting.ssrf.mode.blacklist_description')}
          </Message>

          <Form.Group widths={2}>
            <Form.Input
              label={t('setting.ssrf.domain_list.add_placeholder')}
              placeholder={t('setting.ssrf.domain_list.input_placeholder')}
              value={domainInput}
              onChange={(e, { value }) => setDomainInput(value)}
              onKeyPress={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  addDomainToList();
                }
              }}
              action={
                <Button type='button' onClick={addDomainToList}>
                  {t('setting.ssrf.buttons.add')}
                </Button>
              }
            />
          </Form.Group>

          <div style={{ marginBottom: '1em' }}>
            {domainList.map((domain) => (
              <Label key={domain} style={{ margin: '2px' }}>
                {domain}
                <Label.Detail
                  as='a'
                  onClick={() => removeDomainFromList(domain)}
                >
                  ×
                </Label.Detail>
              </Label>
            ))}
            {domainList.length === 0 && (
              <span style={{ color: '#999' }}>
                {t('setting.ssrf.list_empty')}
              </span>
            )}
          </div>

          <Form.Button type='button' onClick={saveDomainSettings}>
            {t('setting.ssrf.buttons.save_domain')}
          </Form.Button>

          <Divider />

          {/* IP List */}
          <Header as='h4'>{t('setting.ssrf.ip_list.title')}</Header>
          <Message size='small'>{t('setting.ssrf.ip_list.description')}</Message>

          {/* Mode Toggle */}
          <Form.Group inline>
            <label>{t('setting.ssrf.mode_label')}</label>
            <Form.Field>
              <Radio
                label={t('setting.ssrf.mode.blacklist')}
                name='ipListMode'
                value='blacklist'
                checked={ipListMode === 'blacklist'}
                onChange={handleIPModeChange}
              />
            </Form.Field>
            <Form.Field>
              <Radio
                label={t('setting.ssrf.mode.whitelist')}
                name='ipListMode'
                value='whitelist'
                checked={ipListMode === 'whitelist'}
                onChange={handleIPModeChange}
              />
            </Form.Field>
          </Form.Group>

          <Message size='small' color={ipListMode === 'whitelist' ? 'yellow' : undefined}>
            {ipListMode === 'whitelist'
              ? t('setting.ssrf.mode.whitelist_ip_description')
              : t('setting.ssrf.mode.blacklist_ip_description')}
          </Message>

          <Form.Group widths={2}>
            <Form.Input
              label={t('setting.ssrf.ip_list.add_placeholder')}
              placeholder={t('setting.ssrf.ip_list.input_placeholder')}
              value={ipInput}
              onChange={(e, { value }) => setIPInput(value)}
              onKeyPress={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  addIPToList();
                }
              }}
              action={
                <Button type='button' onClick={addIPToList}>
                  {t('setting.ssrf.buttons.add')}
                </Button>
              }
            />
          </Form.Group>

          <div style={{ marginBottom: '1em' }}>
            {ipList.map((ip) => (
              <Label key={ip} style={{ margin: '2px' }}>
                {ip}
                <Label.Detail as='a' onClick={() => removeIPFromList(ip)}>
                  ×
                </Label.Detail>
              </Label>
            ))}
            {ipList.length === 0 && (
              <span style={{ color: '#999' }}>
                {t('setting.ssrf.list_empty')}
              </span>
            )}
          </div>

          <Form.Button type='button' onClick={saveIPSettings}>
            {t('setting.ssrf.buttons.save_ip')}
          </Form.Button>

          <Divider />

          {/* Allowed Ports */}
          <Header as='h4'>{t('setting.ssrf.ports.title')}</Header>
          <Message size='small'>{t('setting.ssrf.ports.description')}</Message>

          <Form.Group widths={2}>
            <Form.Input
              label={t('setting.ssrf.ports.add_placeholder')}
              placeholder={t('setting.ssrf.ports.input_placeholder')}
              value={portInput}
              onChange={(e, { value }) => setPortInput(value)}
              onKeyPress={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  addPort();
                }
              }}
              action={
                <Button type='button' onClick={addPort}>
                  {t('setting.ssrf.buttons.add')}
                </Button>
              }
            />
          </Form.Group>

          <div style={{ marginBottom: '1em' }}>
            {allowedPorts.map((port) => (
              <Label key={port} style={{ margin: '2px' }}>
                {port}
                <Label.Detail as='a' onClick={() => removePort(port)}>
                  ×
                </Label.Detail>
              </Label>
            ))}
          </div>
          <Message size='small'>{t('setting.ssrf.ports.config_description')}</Message>

          <Form.Button type='button' onClick={saveAllowedPorts}>
            {t('setting.ssrf.buttons.save_ports')}
          </Form.Button>
        </Form>
      </Grid.Column>
    </Grid>
  );
};

export default SSRFSetting;