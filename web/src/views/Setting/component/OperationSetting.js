import { useState, useEffect, useContext } from 'react';
import SubCard from 'ui-component/cards/SubCard';
import { Stack, FormControl, InputLabel, OutlinedInput, Checkbox, Button, FormControlLabel, TextField, Alert } from '@mui/material';
import { showSuccess, showError, verifyJSON } from 'utils/common';
import { API } from 'utils/api';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { DateTimePicker } from '@mui/x-date-pickers/DateTimePicker';
import ChatLinksDataGrid from './ChatLinksDataGrid';
import dayjs from 'dayjs';
import { LoadStatusContext } from 'contexts/StatusContext';
require('dayjs/locale/zh-cn');

const OperationSetting = () => {
  let now = new Date();
  let [inputs, setInputs] = useState({
    QuotaForNewUser: 0,
    QuotaForInviter: 0,
    QuotaForInvitee: 0,
    QuotaRemindThreshold: 0,
    PreConsumedQuota: 0,
    GroupRatio: '',
    TopUpLink: '',
    ChatLink: '',
    ChatLinks: '',
    QuotaPerUnit: 0,
    AutomaticDisableChannelEnabled: '',
    AutomaticEnableChannelEnabled: '',
    ChannelDisableThreshold: 0,
    LogConsumeEnabled: '',
    DisplayInCurrencyEnabled: '',
    DisplayTokenStatEnabled: '',
    ApproximateTokenEnabled: '',
    RetryTimes: 0,
    RetryCooldownSeconds: 0,
    MjNotifyEnabled: '',
    ChatCacheEnabled: '',
    ChatCacheExpireMinute: 5,
    ChatImageRequestProxy: ''
  });
  const [originInputs, setOriginInputs] = useState({});
  let [loading, setLoading] = useState(false);
  let [historyTimestamp, setHistoryTimestamp] = useState(now.getTime() / 1000 - 30 * 24 * 3600); // a month ago new Date().getTime() / 1000 + 3600
  const loadStatus = useContext(LoadStatusContext);

  const getOptions = async () => {
    try {
      const res = await API.get('/api/option/');
      const { success, message, data } = res.data;
      if (success) {
        let newInputs = {};
        data.forEach((item) => {
          if (item.key === 'GroupRatio') {
            item.value = JSON.stringify(JSON.parse(item.value), null, 2);
          }
          newInputs[item.key] = item.value;
        });
        setInputs(newInputs);
        setOriginInputs(newInputs);
      } else {
        showError(message);
      }
    } catch (error) {
      return;
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

    try {
      const res = await API.put('/api/option/', {
        key,
        value
      });
      const { success, message } = res.data;
      if (success) {
        setInputs((inputs) => ({ ...inputs, [key]: value }));
        getOptions();
        await loadStatus();
      } else {
        showError(message);
      }
    } catch (error) {
      return;
    }

    setLoading(false);
  };

  const handleInputChange = async (event) => {
    let { name, value } = event.target;

    if (name.endsWith('Enabled')) {
      await updateOption(name, value);
      showSuccess('设置成功！');
    } else {
      setInputs((inputs) => ({ ...inputs, [name]: value }));
    }
  };

  const submitConfig = async (group) => {
    switch (group) {
      case 'monitor':
        if (originInputs['ChannelDisableThreshold'] !== inputs.ChannelDisableThreshold) {
          await updateOption('ChannelDisableThreshold', inputs.ChannelDisableThreshold);
        }
        if (originInputs['QuotaRemindThreshold'] !== inputs.QuotaRemindThreshold) {
          await updateOption('QuotaRemindThreshold', inputs.QuotaRemindThreshold);
        }
        break;
      case 'ratio':
        if (originInputs['GroupRatio'] !== inputs.GroupRatio) {
          if (!verifyJSON(inputs.GroupRatio)) {
            showError('分组倍率不是合法的 JSON 字符串');
            return;
          }
          await updateOption('GroupRatio', inputs.GroupRatio);
        }
        break;
      case 'chatlinks':
        if (originInputs['ChatLinks'] !== inputs.ChatLinks) {
          if (!verifyJSON(inputs.ChatLinks)) {
            showError('links不是合法的 JSON 字符串');
            return;
          }
          await updateOption('ChatLinks', inputs.ChatLinks);
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
        if (inputs.QuotaPerUnit < 0 || inputs.RetryTimes < 0 || inputs.RetryCooldownSeconds < 0) {
          showError('单位额度、重试次数、冷却时间不能为负数');
          return;
        }

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
        if (originInputs['RetryCooldownSeconds'] !== inputs.RetryCooldownSeconds) {
          await updateOption('RetryCooldownSeconds', inputs.RetryCooldownSeconds);
        }
        break;
      case 'other':
        if (originInputs['ChatCacheExpireMinute'] !== inputs.ChatCacheExpireMinute) {
          await updateOption('ChatCacheExpireMinute', inputs.ChatCacheExpireMinute);
        }
        if (originInputs['ChatImageRequestProxy'] !== inputs.ChatImageRequestProxy) {
          await updateOption('ChatImageRequestProxy', inputs.ChatImageRequestProxy);
        }
        break;
    }

    showSuccess('保存成功！');
  };

  const deleteHistoryLogs = async () => {
    try {
      const res = await API.delete(`/api/log/?target_timestamp=${Math.floor(historyTimestamp)}`);
      const { success, message, data } = res.data;
      if (success) {
        showSuccess(`${data} 条日志已清理！`);
        return;
      }
      showError('日志清理失败：' + message);
    } catch (error) {
      return;
    }
  };

  return (
    <Stack spacing={2}>
      <SubCard title="通用设置">
        <Stack justifyContent="flex-start" alignItems="flex-start" spacing={2}>
          <Stack direction={{ sm: 'column', md: 'row' }} spacing={{ xs: 3, sm: 2, md: 4 }}>
            <FormControl fullWidth>
              <InputLabel htmlFor="TopUpLink">充值链接</InputLabel>
              <OutlinedInput
                id="TopUpLink"
                name="TopUpLink"
                value={inputs.TopUpLink}
                onChange={handleInputChange}
                label="充值链接"
                placeholder="例如发卡网站的购买链接"
                disabled={loading}
              />
            </FormControl>
            <FormControl fullWidth>
              <InputLabel htmlFor="ChatLink">聊天链接</InputLabel>
              <OutlinedInput
                id="ChatLink"
                name="ChatLink"
                value={inputs.ChatLink}
                onChange={handleInputChange}
                label="聊天链接"
                placeholder="例如 ChatGPT Next Web 的部署地址"
                disabled={loading}
              />
            </FormControl>
            <FormControl fullWidth>
              <InputLabel htmlFor="QuotaPerUnit">单位额度</InputLabel>
              <OutlinedInput
                id="QuotaPerUnit"
                name="QuotaPerUnit"
                value={inputs.QuotaPerUnit}
                onChange={handleInputChange}
                label="单位额度"
                placeholder="一单位货币能兑换的额度"
                disabled={loading}
              />
            </FormControl>
            <FormControl fullWidth>
              <InputLabel htmlFor="RetryTimes">重试次数</InputLabel>
              <OutlinedInput
                id="RetryTimes"
                name="RetryTimes"
                value={inputs.RetryTimes}
                onChange={handleInputChange}
                label="重试次数"
                placeholder="重试次数"
                disabled={loading}
              />
            </FormControl>
            <FormControl fullWidth>
              <InputLabel htmlFor="RetryCooldownSeconds">重试间隔(秒)</InputLabel>
              <OutlinedInput
                id="RetryCooldownSeconds"
                name="RetryCooldownSeconds"
                value={inputs.RetryCooldownSeconds}
                onChange={handleInputChange}
                label="重试间隔(秒)"
                placeholder="重试间隔(秒)"
                disabled={loading}
              />
            </FormControl>
          </Stack>
          <Stack
            direction={{ sm: 'column', md: 'row' }}
            spacing={{ xs: 3, sm: 2, md: 4 }}
            justifyContent="flex-start"
            alignItems="flex-start"
          >
            <FormControlLabel
              sx={{ marginLeft: '0px' }}
              label="以货币形式显示额度"
              control={
                <Checkbox
                  checked={inputs.DisplayInCurrencyEnabled === 'true'}
                  onChange={handleInputChange}
                  name="DisplayInCurrencyEnabled"
                />
              }
            />

            <FormControlLabel
              label="Billing 相关 API 显示令牌额度而非用户额度"
              control={
                <Checkbox checked={inputs.DisplayTokenStatEnabled === 'true'} onChange={handleInputChange} name="DisplayTokenStatEnabled" />
              }
            />

            <FormControlLabel
              label="使用近似的方式估算 token 数以减少计算量"
              control={
                <Checkbox checked={inputs.ApproximateTokenEnabled === 'true'} onChange={handleInputChange} name="ApproximateTokenEnabled" />
              }
            />
          </Stack>
          <Button
            variant="contained"
            onClick={() => {
              submitConfig('general').then();
            }}
          >
            保存通用设置
          </Button>
        </Stack>
      </SubCard>
      <SubCard title="其他设置">
        <Stack justifyContent="flex-start" alignItems="flex-start" spacing={2}>
          <Stack
            direction={{ sm: 'column', md: 'row' }}
            spacing={{ xs: 3, sm: 2, md: 4 }}
            justifyContent="flex-start"
            alignItems="flex-start"
          >
            <FormControlLabel
              sx={{ marginLeft: '0px' }}
              label="Midjourney 允许回调（会泄露服务器ip地址）"
              control={<Checkbox checked={inputs.MjNotifyEnabled === 'true'} onChange={handleInputChange} name="MjNotifyEnabled" />}
            />
            <FormControlLabel
              sx={{ marginLeft: '0px' }}
              label="是否开启聊天缓存(如果没有启用Redis，将会存储在数据库中)"
              control={<Checkbox checked={inputs.ChatCacheEnabled === 'true'} onChange={handleInputChange} name="ChatCacheEnabled" />}
            />
          </Stack>
          <Stack direction={{ sm: 'column', md: 'row' }} spacing={{ xs: 3, sm: 2, md: 4 }}>
            <FormControl>
              <InputLabel htmlFor="ChatCacheExpireMinute">缓存时间(分钟)</InputLabel>
              <OutlinedInput
                id="ChatCacheExpireMinute"
                name="ChatCacheExpireMinute"
                value={inputs.ChatCacheExpireMinute}
                onChange={handleInputChange}
                label="缓存时间(分钟)"
                placeholder="开启缓存时，数据缓存的时间"
                disabled={loading}
              />
            </FormControl>
          </Stack>

          <Stack spacing={2}>
            <Alert severity="info">
              当用户使用vision模型并提供了图片链接时，我们的服务器需要下载这些图片并计算 tokens。为了在下载图片时保护服务器的 IP
              地址不被泄露，可以在下方配置一个代理。这个代理配置使用的是 HTTP 或 SOCKS5
              代理。如果你是个人用户，这个配置可以不用理会。代理格式为 http://127.0.0.1:1080 或 socks5://127.0.0.1:1080
            </Alert>
            <FormControl>
              <InputLabel htmlFor="ChatImageRequestProxy">图片检测代理</InputLabel>
              <OutlinedInput
                id="ChatImageRequestProxy"
                name="ChatImageRequestProxy"
                value={inputs.ChatImageRequestProxy}
                onChange={handleInputChange}
                label="图片检测代理"
                placeholder="聊天图片检测代理设置，如果不设置可能会泄漏服务器ip"
                disabled={loading}
              />
            </FormControl>
          </Stack>
          <Button
            variant="contained"
            onClick={() => {
              submitConfig('other').then();
            }}
          >
            保存其他设置
          </Button>
        </Stack>
      </SubCard>
      <SubCard title="日志设置">
        <Stack direction="column" justifyContent="flex-start" alignItems="flex-start" spacing={2}>
          <FormControlLabel
            label="启用日志消费"
            control={<Checkbox checked={inputs.LogConsumeEnabled === 'true'} onChange={handleInputChange} name="LogConsumeEnabled" />}
          />

          <FormControl>
            <LocalizationProvider dateAdapter={AdapterDayjs} adapterLocale={'zh-cn'}>
              <DateTimePicker
                label="日志清理时间"
                placeholder="日志清理时间"
                ampm={false}
                name="historyTimestamp"
                value={historyTimestamp === null ? null : dayjs.unix(historyTimestamp)}
                disabled={loading}
                onChange={(newValue) => {
                  setHistoryTimestamp(newValue === null ? null : newValue.unix());
                }}
                slotProps={{
                  actionBar: {
                    actions: ['today', 'clear', 'accept']
                  }
                }}
              />
            </LocalizationProvider>
          </FormControl>
          <Button
            variant="contained"
            onClick={() => {
              deleteHistoryLogs().then();
            }}
          >
            清理历史日志
          </Button>
        </Stack>
      </SubCard>
      <SubCard title="监控设置">
        <Stack justifyContent="flex-start" alignItems="flex-start" spacing={2}>
          <Stack direction={{ sm: 'column', md: 'row' }} spacing={{ xs: 3, sm: 2, md: 4 }}>
            <FormControl fullWidth>
              <InputLabel htmlFor="ChannelDisableThreshold">最长响应时间</InputLabel>
              <OutlinedInput
                id="ChannelDisableThreshold"
                name="ChannelDisableThreshold"
                type="number"
                value={inputs.ChannelDisableThreshold}
                onChange={handleInputChange}
                label="最长响应时间"
                placeholder="单位秒，当运行通道全部测试时，超过此时间将自动禁用通道"
                disabled={loading}
              />
            </FormControl>
            <FormControl fullWidth>
              <InputLabel htmlFor="QuotaRemindThreshold">额度提醒阈值</InputLabel>
              <OutlinedInput
                id="QuotaRemindThreshold"
                name="QuotaRemindThreshold"
                type="number"
                value={inputs.QuotaRemindThreshold}
                onChange={handleInputChange}
                label="额度提醒阈值"
                placeholder="低于此额度时将发送邮件提醒用户"
                disabled={loading}
              />
            </FormControl>
          </Stack>
          <FormControlLabel
            label="失败时自动禁用通道"
            control={
              <Checkbox
                checked={inputs.AutomaticDisableChannelEnabled === 'true'}
                onChange={handleInputChange}
                name="AutomaticDisableChannelEnabled"
              />
            }
          />
          <FormControlLabel
            label="成功时自动启用通道"
            control={
              <Checkbox
                checked={inputs.AutomaticEnableChannelEnabled === 'true'}
                onChange={handleInputChange}
                name="AutomaticEnableChannelEnabled"
              />
            }
          />
          <Button
            variant="contained"
            onClick={() => {
              submitConfig('monitor').then();
            }}
          >
            保存监控设置
          </Button>
        </Stack>
      </SubCard>
      <SubCard title="额度设置">
        <Stack justifyContent="flex-start" alignItems="flex-start" spacing={2}>
          <Stack direction={{ sm: 'column', md: 'row' }} spacing={{ xs: 3, sm: 2, md: 4 }}>
            <FormControl fullWidth>
              <InputLabel htmlFor="QuotaForNewUser">新用户初始额度</InputLabel>
              <OutlinedInput
                id="QuotaForNewUser"
                name="QuotaForNewUser"
                type="number"
                value={inputs.QuotaForNewUser}
                onChange={handleInputChange}
                label="新用户初始额度"
                placeholder="例如：100"
                disabled={loading}
              />
            </FormControl>
            <FormControl fullWidth>
              <InputLabel htmlFor="PreConsumedQuota">请求预扣费额度</InputLabel>
              <OutlinedInput
                id="PreConsumedQuota"
                name="PreConsumedQuota"
                type="number"
                value={inputs.PreConsumedQuota}
                onChange={handleInputChange}
                label="请求预扣费额度"
                placeholder="请求结束后多退少补"
                disabled={loading}
              />
            </FormControl>
            <FormControl fullWidth>
              <InputLabel htmlFor="QuotaForInviter">邀请新用户奖励额度</InputLabel>
              <OutlinedInput
                id="QuotaForInviter"
                name="QuotaForInviter"
                type="number"
                label="邀请新用户奖励额度"
                value={inputs.QuotaForInviter}
                onChange={handleInputChange}
                placeholder="例如：2000"
                disabled={loading}
              />
            </FormControl>
            <FormControl fullWidth>
              <InputLabel htmlFor="QuotaForInvitee">新用户使用邀请码奖励额度</InputLabel>
              <OutlinedInput
                id="QuotaForInvitee"
                name="QuotaForInvitee"
                type="number"
                label="新用户使用邀请码奖励额度"
                value={inputs.QuotaForInvitee}
                onChange={handleInputChange}
                autoComplete="new-password"
                placeholder="例如：1000"
                disabled={loading}
              />
            </FormControl>
          </Stack>
          <Button
            variant="contained"
            onClick={() => {
              submitConfig('quota').then();
            }}
          >
            保存额度设置
          </Button>
        </Stack>
      </SubCard>
      <SubCard title="倍率设置">
        <Stack justifyContent="flex-start" alignItems="flex-start" spacing={2}>
          <FormControl fullWidth>
            <TextField
              multiline
              maxRows={15}
              id="channel-GroupRatio-label"
              label="分组倍率"
              value={inputs.GroupRatio}
              name="GroupRatio"
              onChange={handleInputChange}
              aria-describedby="helper-text-channel-GroupRatio-label"
              minRows={5}
              placeholder="为一个 JSON 文本，键为分组名称，值为倍率"
            />
          </FormControl>

          <Button
            variant="contained"
            onClick={() => {
              submitConfig('ratio').then();
            }}
          >
            保存倍率设置
          </Button>
        </Stack>
      </SubCard>

      <SubCard title="聊天链接设置">
        <Stack spacing={2}>
          <Alert severity="info">
            配置聊天链接，该配置在令牌中的聊天生效以及首页的Playground中的聊天生效. <br />
            链接中可以使{'{key}'}替换用户的令牌，{'{server}'}替换服务器地址。例如：
            {'https://chat.oneapi.pro/#/?settings={"key":"sk-{key}","url":"{server}"}'}
            <br />
            如果未配置，会默认配置以下4个链接：
            <br />
            ChatGPT Next ： {'https://chat.oneapi.pro/#/?settings={"key":"{key}","url":"{server}"}'}
            <br />
            chatgpt-web-midjourney-proxy ： {'https://vercel.ddaiai.com/#/?settings={"key":"{key}","url":"{server}"}'}
            <br />
            AMA 问天 ： {'ama://set-api-key?server={server}&key={key}'}
            <br />
            opencat ： {'opencat://team/join?domain={server}&token={key}'}
            <br />
            排序规则：值越大越靠前，值相同则按照配置顺序
          </Alert>
          <Stack justifyContent="flex-start" alignItems="flex-start" spacing={2}>
            <ChatLinksDataGrid links={inputs.ChatLinks || '[]'} onChange={handleInputChange} />

            <Button
              variant="contained"
              onClick={() => {
                submitConfig('chatlinks').then();
              }}
            >
              保存聊天链接设置
            </Button>
          </Stack>
        </Stack>
      </SubCard>
    </Stack>
  );
};

export default OperationSetting;
