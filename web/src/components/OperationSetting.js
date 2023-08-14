import React, {useEffect, useState} from 'react';
import {Divider, Form, Grid, Header} from 'semantic-ui-react';
import {API, showError, verifyJSON} from '../helpers';

const OperationSetting = () => {
    let [inputs, setInputs] = useState({
        QuotaForNewUser: 0,
        QuotaForInviter: 0,
        QuotaForInvitee: 0,
        QuotaRemindThreshold: 0,
        PreConsumedQuota: 0,
        ModelRatio: '',
        GroupRatio: '',
        TopUpLink: '',
        ChatLink: '',
        QuotaPerUnit: 0,
        AutomaticDisableChannelEnabled: '',
        ChannelDisableThreshold: 0,
        LogConsumeEnabled: '',
        DisplayInCurrencyEnabled: '',
        DisplayTokenStatEnabled: '',
        ApproximateTokenEnabled: '',
        RetryTimes: 0,
        StablePrice: 6,
        NormalPrice: 1.5,
        BasePrice: 1.5,
    });
    const [originInputs, setOriginInputs] = useState({});
    let [loading, setLoading] = useState(false);

    const getOptions = async () => {
        const res = await API.get('/api/option/');
        const {success, message, data} = res.data;
        if (success) {
            let newInputs = {};
            data.forEach((item) => {
                if (item.key === 'ModelRatio' || item.key === 'GroupRatio') {
                    item.value = JSON.stringify(JSON.parse(item.value), null, 2);
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
            value
        });
        const {success, message} = res.data;
        if (success) {
            setInputs((inputs) => ({...inputs, [key]: value}));
        } else {
            showError(message);
        }
        setLoading(false);
    };

    const handleInputChange = async (e, {name, value}) => {
        if (name.endsWith('Enabled')) {
            await updateOption(name, value);
        } else {
            setInputs((inputs) => ({...inputs, [name]: value}));
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
            case 'stable':
                await updateOption('StablePrice', inputs.StablePrice);
                await updateOption('NormalPrice', inputs.NormalPrice);
                await updateOption('BasePrice', inputs.BasePrice);
                localStorage.setItem('stable_price', inputs.StablePrice);
                localStorage.setItem('normal_price', inputs.NormalPrice);
                localStorage.setItem('base_price', inputs.BasePrice);
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

    return (
        <Grid columns={1}>
            <Grid.Column>
                <Form loading={loading}>
                    <Header as='h3'>
                        通用设置
                    </Header>
                    <Form.Group widths={4}>
                        <Form.Input
                            label='充值链接'
                            name='TopUpLink'
                            onChange={handleInputChange}
                            autoComplete='new-password'
                            value={inputs.TopUpLink}
                            type='link'
                            placeholder='例如发卡网站的购买链接'
                        />
                        <Form.Input
                            label='聊天页面链接'
                            name='ChatLink'
                            onChange={handleInputChange}
                            autoComplete='new-password'
                            value={inputs.ChatLink}
                            type='link'
                            placeholder='例如 ChatGPT Next Web 的部署地址'
                        />
                        <Form.Input
                            label='单位美元额度'
                            name='QuotaPerUnit'
                            onChange={handleInputChange}
                            autoComplete='new-password'
                            value={inputs.QuotaPerUnit}
                            type='number'
                            step='0.01'
                            placeholder='一单位货币能兑换的额度'
                        />
                        <Form.Input
                            label='失败重试次数'
                            name='RetryTimes'
                            type={'number'}
                            step='1'
                            min='0'
                            onChange={handleInputChange}
                            autoComplete='new-password'
                            value={inputs.RetryTimes}
                            placeholder='失败重试次数'
                        />
                    </Form.Group>
                    <Form.Group inline>
                        <Form.Checkbox
                            checked={inputs.LogConsumeEnabled === 'true'}
                            label='启用额度消费日志记录'
                            name='LogConsumeEnabled'
                            onChange={handleInputChange}
                        />
                        <Form.Checkbox
                            checked={inputs.DisplayInCurrencyEnabled === 'true'}
                            label='以货币形式显示额度'
                            name='DisplayInCurrencyEnabled'
                            onChange={handleInputChange}
                        />
                        <Form.Checkbox
                            checked={inputs.DisplayTokenStatEnabled === 'true'}
                            label='Billing 相关 API 显示令牌额度而非用户额度'
                            name='DisplayTokenStatEnabled'
                            onChange={handleInputChange}
                        />
                        <Form.Checkbox
                            checked={inputs.ApproximateTokenEnabled === 'true'}
                            label='使用近似的方式估算 token 数以减少计算量'
                            name='ApproximateTokenEnabled'
                            onChange={handleInputChange}
                        />
                    </Form.Group>
                    <Form.Button onClick={() => {
                        submitConfig('general').then();
                    }}>保存通用设置</Form.Button>
                    <Divider/>
                    <Header as='h3'>
                        监控设置
                    </Header>
                    <Form.Group widths={3}>
                        <Form.Input
                            label='最长响应时间'
                            name='ChannelDisableThreshold'
                            onChange={handleInputChange}
                            autoComplete='new-password'
                            value={inputs.ChannelDisableThreshold}
                            type='number'
                            min='0'
                            placeholder='单位秒，当运行通道全部测试时，超过此时间将自动禁用通道'
                        />
                        <Form.Input
                            label='额度提醒阈值'
                            name='QuotaRemindThreshold'
                            onChange={handleInputChange}
                            autoComplete='new-password'
                            value={inputs.QuotaRemindThreshold}
                            type='number'
                            min='0'
                            placeholder='低于此额度时将发送邮件提醒用户'
                        />
                    </Form.Group>
                    <Form.Group inline>
                        <Form.Checkbox
                            checked={inputs.AutomaticDisableChannelEnabled === 'true'}
                            label='失败时自动禁用通道'
                            name='AutomaticDisableChannelEnabled'
                            onChange={handleInputChange}
                        />
                    </Form.Group>
                    <Form.Button onClick={() => {
                        submitConfig('monitor').then();
                    }}>保存监控设置</Form.Button>
                    <Divider/>
                    <Header as='h3'>
                        通道设置
                    </Header>
                    <Form.Group widths={3}>
                        <Form.Input
                            label='普通渠道价格'
                            name='NormalPrice'
                            onChange={handleInputChange}
                            autoComplete='new-password'
                            value={inputs.NormalPrice}
                            type='number'
                            // min='1.5'
                            placeholder='n元/刀'
                        />
                        <Form.Input
                            label='稳定渠道价格'
                            name='StablePrice'
                            onChange={handleInputChange}
                            autoComplete='new-password'
                            value={inputs.StablePrice}
                            type='number'
                            // min='1.5'
                            placeholder='n元/刀'
                        />
                    </Form.Group>
                    <Form.Button onClick={() => {
                        submitConfig('stable').then();
                    }}>保存通道设置</Form.Button>
                    <Divider/>
                    <Header as='h3'>
                        额度设置
                    </Header>
                    <Form.Group widths={4}>
                        <Form.Input
                            label='新用户初始额度'
                            name='QuotaForNewUser'
                            onChange={handleInputChange}
                            autoComplete='new-password'
                            value={inputs.QuotaForNewUser}
                            type='number'
                            min='0'
                            placeholder='例如：100'
                        />
                        <Form.Input
                            label='请求预扣费额度'
                            name='PreConsumedQuota'
                            onChange={handleInputChange}
                            autoComplete='new-password'
                            value={inputs.PreConsumedQuota}
                            type='number'
                            min='0'
                            placeholder='请求结束后多退少补'
                        />
                        <Form.Input
                            label='邀请新用户奖励额度'
                            name='QuotaForInviter'
                            onChange={handleInputChange}
                            autoComplete='new-password'
                            value={inputs.QuotaForInviter}
                            type='number'
                            min='0'
                            placeholder='例如：2000'
                        />
                        <Form.Input
                            label='新用户使用邀请码奖励额度'
                            name='QuotaForInvitee'
                            onChange={handleInputChange}
                            autoComplete='new-password'
                            value={inputs.QuotaForInvitee}
                            type='number'
                            min='0'
                            placeholder='例如：1000'
                        />
                    </Form.Group>
                    <Form.Button onClick={() => {
                        submitConfig('quota').then();
                    }}>保存额度设置</Form.Button>
                    <Divider/>
                    <Header as='h3'>
                        倍率设置
                    </Header>
                    <Form.Group widths='equal'>
                        <Form.TextArea
                            label='模型倍率'
                            name='ModelRatio'
                            onChange={handleInputChange}
                            style={{minHeight: 250, fontFamily: 'JetBrains Mono, Consolas'}}
                            autoComplete='new-password'
                            value={inputs.ModelRatio}
                            placeholder='为一个 JSON 文本，键为模型名称，值为倍率'
                        />
                    </Form.Group>
                    <Form.Group widths='equal'>
                        <Form.TextArea
                            label='分组倍率'
                            name='GroupRatio'
                            onChange={handleInputChange}
                            style={{minHeight: 250, fontFamily: 'JetBrains Mono, Consolas'}}
                            autoComplete='new-password'
                            value={inputs.GroupRatio}
                            placeholder='为一个 JSON 文本，键为分组名称，值为倍率'
                        />
                    </Form.Group>
                    <Form.Button onClick={() => {
                        submitConfig('ratio').then();
                    }}>保存倍率设置</Form.Button>
                </Form>
            </Grid.Column>
        </Grid>
    )
        ;
};

export default OperationSetting;
