import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Divider, Form, Grid, Header } from 'semantic-ui-react';
import {
  API,
  showError,
  showSuccess,
  timestamp2string,
  verifyJSON,
} from '../helpers';

const OperationSetting = () => {
  const { t } = useTranslation();
  let now = new Date();
  let [inputs, setInputs] = useState({
    QuotaForNewUser: 0,
    QuotaForInviter: 0,
    QuotaForInvitee: 0,
    QuotaRemindThreshold: 0,
    PreConsumedQuota: 0,
    ModelRatio: '',
    CompletionRatio: '',
    GroupRatio: '',
    TopUpLink: '',
    ChatLink: '',
    QuotaPerUnit: 0,
    AutomaticDisableChannelEnabled: '',
    AutomaticEnableChannelEnabled: '',
    ChannelDisableThreshold: 0,
    LogConsumeEnabled: '',
    DisplayInCurrencyEnabled: '',
    DisplayTokenStatEnabled: '',
    ApproximateTokenEnabled: '',
    RetryTimes: 0,
  });
  const [originInputs, setOriginInputs] = useState({});
  let [loading, setLoading] = useState(false);
  const [batchRatio, setBatchRatio] = useState({
    models: '',
    input: '',
    output: '',
    group: '',
  });
  const [batchRatioLoading, setBatchRatioLoading] = useState(false);
  let [historyTimestamp, setHistoryTimestamp] = useState(
    timestamp2string(now.getTime() / 1000 - 30 * 24 * 3600)
  ); // a month ago

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        if (
          item.key === 'ModelRatio' ||
          item.key === 'GroupRatio' ||
          item.key === 'CompletionRatio'
        ) {
          item.value = JSON.stringify(JSON.parse(item.value), null, 2);
        }
        if (item.value === '{}') {
          item.value = '';
        }
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
    if (key.endsWith('Enabled')) {
      value = inputs[key] === 'true' ? 'false' : 'true';
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
    if (name.endsWith('Enabled')) {
      await updateOption(name, value);
    } else {
      setInputs((inputs) => ({ ...inputs, [name]: value }));
    }
  };

  const submitConfig = async (group) => {
    switch (group) {
      case 'monitor':
        if (
          originInputs['ChannelDisableThreshold'] !==
          inputs.ChannelDisableThreshold
        ) {
          await updateOption(
            'ChannelDisableThreshold',
            inputs.ChannelDisableThreshold
          );
        }
        if (
          originInputs['QuotaRemindThreshold'] !== inputs.QuotaRemindThreshold
        ) {
          await updateOption(
            'QuotaRemindThreshold',
            inputs.QuotaRemindThreshold
          );
        }
        break;
      case 'ratio':
        if (originInputs['ModelRatio'] !== inputs.ModelRatio) {
          if (!verifyJSON(inputs.ModelRatio)) {
            showError('模型倍率不是合法的 JSON 字符串');
            return;
          }
          await updateOption('ModelRatio', inputs.ModelRatio);
        }
        if (originInputs['GroupRatio'] !== inputs.GroupRatio) {
          if (!verifyJSON(inputs.GroupRatio)) {
            showError('分组倍率不是合法的 JSON 字符串');
            return;
          }
          await updateOption('GroupRatio', inputs.GroupRatio);
        }
        if (originInputs['CompletionRatio'] !== inputs.CompletionRatio) {
          if (!verifyJSON(inputs.CompletionRatio)) {
            showError('补全倍率不是合法的 JSON 字符串');
            return;
          }
          await updateOption('CompletionRatio', inputs.CompletionRatio);
        }
        break;
      case 'quota':
        if (originInputs['QuotaForNewUser'] !== inputs.QuotaForNewUser) {
          await updateOption('QuotaForNewUser', inputs.QuotaForNewUser);
        }
        if (originInputs['QuotaForInvitee'] !== inputs.QuotaForInvitee) {
          await updateOption('QuotaForInvitee', inputs.QuotaForInvitee);
        }
        if (originInputs['QuotaForInviter'] !== inputs.QuotaForInviter) {
          await updateOption('QuotaForInviter', inputs.QuotaForInviter);
        }
        if (originInputs['PreConsumedQuota'] !== inputs.PreConsumedQuota) {
          await updateOption('PreConsumedQuota', inputs.PreConsumedQuota);
        }
        break;
      case 'general':
        if (originInputs['TopUpLink'] !== inputs.TopUpLink) {
          await updateOption('TopUpLink', inputs.TopUpLink);
        }
        if (originInputs['ChatLink'] !== inputs.ChatLink) {
          await updateOption('ChatLink', inputs.ChatLink);
        }
        if (originInputs['QuotaPerUnit'] !== inputs.QuotaPerUnit) {
          await updateOption('QuotaPerUnit', inputs.QuotaPerUnit);
        }
        if (originInputs['RetryTimes'] !== inputs.RetryTimes) {
          await updateOption('RetryTimes', inputs.RetryTimes);
        }
        break;
    }
  };

  const deleteHistoryLogs = async () => {
    console.log(inputs);
    const res = await API.delete(
      `/api/log/?target_timestamp=${Date.parse(historyTimestamp) / 1000}`
    );
    const { success, message, data } = res.data;
    if (success) {
      showSuccess(`${data} 条日志已清理！`);
      return;
    }
    showError('日志清理失败：' + message);
  };

  const submitBatchRatio = async () => {
    const models = batchRatio.models
      .split(/[\n,]/)
      .map((s) => s.trim())
      .filter((s) => s.length > 0);
    if (models.length === 0) {
      showError(t('setting.operation.ratio.batch.models_required'));
      return;
    }
    const input = parseFloat(batchRatio.input);
    const output =
      batchRatio.output === '' ? input : parseFloat(batchRatio.output);
    if (!(input > 0)) {
      showError(t('setting.operation.ratio.batch.input_required'));
      return;
    }
    if (!(output >= 0)) {
      showError(t('setting.operation.ratio.batch.output_invalid'));
      return;
    }
    setBatchRatioLoading(true);
    try {
      const res = await API.post('/api/option/batch_ratio', {
        models,
        input,
        output,
        group: batchRatio.group.trim(),
      });
      const { success, message, data } = res.data;
      if (success) {
        showSuccess(
          t('setting.operation.ratio.batch.success', { count: data })
        );
        setBatchRatio({ models: '', input: '', output: '', group: '' });
        await getOptions();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setBatchRatioLoading(false);
  };

  return (
    <Grid columns={1}>
      <Grid.Column>
        <Form loading={loading}>
          <Header as='h3'>{t('setting.operation.quota.title')}</Header>
          <Form.Group widths='equal'>
            <Form.Input
              label={t('setting.operation.quota.new_user')}
              name='QuotaForNewUser'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.QuotaForNewUser}
              type='number'
              min='0'
              placeholder={t('setting.operation.quota.new_user_placeholder')}
            />
            <Form.Input
              label={t('setting.operation.quota.pre_consume')}
              name='PreConsumedQuota'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.PreConsumedQuota}
              type='number'
              min='0'
              placeholder={t('setting.operation.quota.pre_consume_placeholder')}
            />
            <Form.Input
              label={t('setting.operation.quota.inviter_reward')}
              name='QuotaForInviter'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.QuotaForInviter}
              type='number'
              min='0'
              placeholder={t(
                'setting.operation.quota.inviter_reward_placeholder'
              )}
            />
            <Form.Input
              label={t('setting.operation.quota.invitee_reward')}
              name='QuotaForInvitee'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.QuotaForInvitee}
              type='number'
              min='0'
              placeholder={t(
                'setting.operation.quota.invitee_reward_placeholder'
              )}
            />
          </Form.Group>
          <Form.Button
            onClick={() => {
              submitConfig('quota').then();
            }}
          >
            {t('setting.operation.quota.buttons.save')}
          </Form.Button>
          <Divider />
          <Header as='h3'>{t('setting.operation.ratio.title')}</Header>
          <Form.Group widths='equal'>
            <Form.TextArea
              label={t('setting.operation.ratio.model.title')}
              name='ModelRatio'
              onChange={handleInputChange}
              style={{ minHeight: 250, fontFamily: 'JetBrains Mono, Consolas' }}
              autoComplete='new-password'
              value={inputs.ModelRatio}
              placeholder={t('setting.operation.ratio.model.placeholder')}
            />
          </Form.Group>
          <Form.Group widths='equal'>
            <Form.TextArea
              label={t('setting.operation.ratio.completion.title')}
              name='CompletionRatio'
              onChange={handleInputChange}
              style={{ minHeight: 250, fontFamily: 'JetBrains Mono, Consolas' }}
              autoComplete='new-password'
              value={inputs.CompletionRatio}
              placeholder={t('setting.operation.ratio.completion.placeholder')}
            />
          </Form.Group>
          <Form.Group widths='equal'>
            <Form.TextArea
              label={t('setting.operation.ratio.group.title')}
              name='GroupRatio'
              onChange={handleInputChange}
              style={{ minHeight: 250, fontFamily: 'JetBrains Mono, Consolas' }}
              autoComplete='new-password'
              value={inputs.GroupRatio}
              placeholder={t('setting.operation.ratio.group.placeholder')}
            />
          </Form.Group>
          <Form.Button
            onClick={() => {
              submitConfig('ratio').then();
            }}
          >
            {t('setting.operation.ratio.buttons.save')}
          </Form.Button>
          <Divider />
          <Header as='h4'>{t('setting.operation.ratio.batch.title')}</Header>
          <Form.Group widths='equal'>
            <Form.TextArea
              label={t('setting.operation.ratio.batch.models')}
              placeholder={t('setting.operation.ratio.batch.models_placeholder')}
              value={batchRatio.models}
              onChange={(e, { value }) =>
                setBatchRatio({ ...batchRatio, models: value })
              }
              style={{ minHeight: 120, fontFamily: 'JetBrains Mono, Consolas' }}
            />
            <div>
              <Form.Input
                fluid
                label={t('setting.operation.ratio.batch.input')}
                type='number'
                step='0.0001'
                min='0'
                placeholder={t('setting.operation.ratio.batch.input_placeholder')}
                value={batchRatio.input}
                onChange={(e, { value }) =>
                  setBatchRatio({ ...batchRatio, input: value })
                }
              />
              <Form.Input
                fluid
                label={t('setting.operation.ratio.batch.output')}
                type='number'
                step='0.0001'
                min='0'
                placeholder={t('setting.operation.ratio.batch.output_placeholder')}
                value={batchRatio.output}
                onChange={(e, { value }) =>
                  setBatchRatio({ ...batchRatio, output: value })
                }
              />
              <Form.Input
                fluid
                label={t('setting.operation.ratio.batch.group')}
                placeholder={t('setting.operation.ratio.batch.group_placeholder')}
                value={batchRatio.group}
                onChange={(e, { value }) =>
                  setBatchRatio({ ...batchRatio, group: value })
                }
              />
            </div>
          </Form.Group>
          <Form.Button loading={batchRatioLoading} onClick={submitBatchRatio}>
            {t('setting.operation.ratio.batch.apply')}
          </Form.Button>
          <Divider />
          <Header as='h3'>{t('setting.operation.log.title')}</Header>
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.LogConsumeEnabled === 'true'}
              label={t('setting.operation.log.enable_consume')}
              name='LogConsumeEnabled'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Group widths={4}>
            <Form.Input
              label={t('setting.operation.log.target_time')}
              value={historyTimestamp}
              type='datetime-local'
              name='history_timestamp'
              onChange={(e, { name, value }) => {
                setHistoryTimestamp(value);
              }}
            />
          </Form.Group>
          <Form.Button
            onClick={() => {
              deleteHistoryLogs().then();
            }}
          >
            {t('setting.operation.log.buttons.clean')}
          </Form.Button>

          <Divider />
          <Header as='h3'>{t('setting.operation.monitor.title')}</Header>
          <Form.Group widths={3}>
            <Form.Input
              label={t('setting.operation.monitor.max_response_time')}
              name='ChannelDisableThreshold'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.ChannelDisableThreshold}
              type='number'
              min='0'
              placeholder={t(
                'setting.operation.monitor.max_response_time_placeholder'
              )}
            />
            <Form.Input
              label={t('setting.operation.monitor.quota_reminder')}
              name='QuotaRemindThreshold'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.QuotaRemindThreshold}
              type='number'
              min='0'
              placeholder={t(
                'setting.operation.monitor.quota_reminder_placeholder'
              )}
            />
          </Form.Group>
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.AutomaticDisableChannelEnabled === 'true'}
              label={t('setting.operation.monitor.auto_disable')}
              name='AutomaticDisableChannelEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.AutomaticEnableChannelEnabled === 'true'}
              label={t('setting.operation.monitor.auto_enable')}
              name='AutomaticEnableChannelEnabled'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Button
            onClick={() => {
              submitConfig('monitor').then();
            }}
          >
            {t('setting.operation.monitor.buttons.save')}
          </Form.Button>

          <Divider />
          <Header as='h3'>{t('setting.operation.general.title')}</Header>
          <Form.Group widths={4}>
            <Form.Input
              label={t('setting.operation.general.topup_link')}
              name='TopUpLink'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.TopUpLink}
              type='link'
              placeholder={t(
                'setting.operation.general.topup_link_placeholder'
              )}
            />
            <Form.Input
              label={t('setting.operation.general.chat_link')}
              name='ChatLink'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.ChatLink}
              type='link'
              placeholder={t('setting.operation.general.chat_link_placeholder')}
            />
            <Form.Input
              label={t('setting.operation.general.quota_per_unit')}
              name='QuotaPerUnit'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.QuotaPerUnit}
              type='number'
              step='0.01'
              placeholder={t(
                'setting.operation.general.quota_per_unit_placeholder'
              )}
            />
            <Form.Input
              label={t('setting.operation.general.retry_times')}
              name='RetryTimes'
              type={'number'}
              step='1'
              min='0'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.RetryTimes}
              placeholder={t(
                'setting.operation.general.retry_times_placeholder'
              )}
            />
          </Form.Group>
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.DisplayInCurrencyEnabled === 'true'}
              label={t('setting.operation.general.display_in_currency')}
              name='DisplayInCurrencyEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.DisplayTokenStatEnabled === 'true'}
              label={t('setting.operation.general.display_token_stat')}
              name='DisplayTokenStatEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.ApproximateTokenEnabled === 'true'}
              label={t('setting.operation.general.approximate_token')}
              name='ApproximateTokenEnabled'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Button
            onClick={() => {
              submitConfig('general').then();
            }}
          >
            {t('setting.operation.general.buttons.save')}
          </Form.Button>
        </Form>
      </Grid.Column>
    </Grid>
  );
};

export default OperationSetting;
