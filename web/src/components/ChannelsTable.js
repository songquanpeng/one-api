import React, { useEffect, useState } from 'react';
import { Button, Form, Input, Label, Message, Pagination, Popup, Table } from 'semantic-ui-react';
import { Link } from 'react-router-dom';
import { API, setPromptShown, shouldShowPrompt, showError, showInfo, showSuccess, timestamp2string } from '../helpers';

import { CHANNEL_OPTIONS, ITEMS_PER_PAGE } from '../constants';
import { renderGroup, renderNumber } from '../helpers/render';

function renderTimestamp(timestamp) {
  return (
    <>
      {timestamp2string(timestamp)}
    </>
  );
}

let type2label = undefined;

function renderType(type) {
  if (!type2label) {
    type2label = new Map;
    for (let i = 0; i < CHANNEL_OPTIONS.length; i++) {
      type2label[CHANNEL_OPTIONS[i].value] = CHANNEL_OPTIONS[i];
    }
    type2label[0] = { value: 0, text: '未知类型', color: 'grey' };
  }
  return <Label basic color={type2label[type]?.color}>{type2label[type]?.text}</Label>;
}

function renderBalance(type, balance) {
  switch (type) {
    case 1: // OpenAI
      return <span style={{ color: 'var(--czl-primary-color)' }}>${balance.toFixed(2)}</span>;
    case 4: // CloseAI
      return <span style={{ color: 'var(--czl-primary-color-hover)' }}>¥{balance.toFixed(2)}</span>;
    case 8: // 自定义
      return <span style={{ color: 'var(--czl-success-color)' }}>${balance.toFixed(2)}</span>;
    case 5: // OpenAI-SB
      return <span style={{ color: 'var(--czl-primary-color-suppl)' }}>¥{(balance / 10000).toFixed(2)}</span>;
    case 10: // AI Proxy
      return <span style={{ color: 'var(--czl-success-color)' }}>{renderNumber(balance)}</span>;
    case 12: // API2GPT
      return <span style={{ color: 'var(--czl-error-color)' }}>¥{balance.toFixed(2)}</span>;
    case 13: // AIGC2D
      return <span style={{ color: 'var(--czl-warning-color)' }}>{renderNumber(balance)}</span>;
    default:
      return <span style={{ color: 'var(--czl-info-color)' }}>不支持</span>;

  }
}

const ChannelsTable = () => {
  const [channels, setChannels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [updatingBalance, setUpdatingBalance] = useState(false);
  const [showPrompt, setShowPrompt] = useState(shouldShowPrompt("channel-test"));
  const [monthlyQuotas, setMonthlyQuotas] = useState({});

  // 获取本月的开始和结束时间戳
  function getMonthStartAndEndTimestamps() {
    const now = new Date();
    const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);
    const endOfMonth = new Date(now.getFullYear(), now.getMonth() + 1, 0, 23, 59, 59, 999); // 设置到月末的最后一刻

    // 将日期转换为UNIX时间戳（秒数）
    const startTimestamp = Math.floor(startOfMonth.getTime() / 1000);
    const endTimestamp = Math.floor(endOfMonth.getTime() / 1000);

    return { startTimestamp, endTimestamp };
  }

  // 获取本月的配额
  const fetchMonthlyQuotasAndChannels = async (fetchedChannels) => {
    const { startTimestamp, endTimestamp } = getMonthStartAndEndTimestamps();
    
    const quotaRequests = fetchedChannels.map(channel => (
      API.get(`/api/log/stat?type=0&start_timestamp=${startTimestamp}&end_timestamp=${endTimestamp}&channel=${channel.id}`)
    ));
  
    try {
      const quotaResponses = await Promise.all(quotaRequests);
      const quotaPerUnit = localStorage.getItem('quota_per_unit')  || 500000;
      const newMonthlyQuotas = quotaResponses.reduce((acc, response, index) => {
        const quota = (response.data.data.quota / quotaPerUnit).toFixed(3);
        const channelId = fetchedChannels[index].id;
        acc[channelId] = parseFloat(quota);
        return acc;
      }, {});
    
      setMonthlyQuotas(newMonthlyQuotas);
    } catch (error) {
      console.error('获取月度配额失败:', error);
    }
  };
  
  // 加载频道列表
  const loadChannels = async (startIdx) => {
    const res = await API.get(`/api/channel/?p=${startIdx}`);
    const { success, message, data } = res.data;
    if (success) {
      const fetchedChannels = data;
  
      if (startIdx === 0) {
        setChannels(fetchedChannels);
      } else {
        let newChannels = [...channels];
        newChannels.splice(startIdx * ITEMS_PER_PAGE, fetchedChannels.length, ...fetchedChannels);
        setChannels(newChannels);
      }
      fetchMonthlyQuotasAndChannels(fetchedChannels); // 在这里调用函数
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(channels.length / ITEMS_PER_PAGE) + 1) {
        // In this case we have to load more data and then append them.
        await loadChannels(activePage - 1);
      }
      setActivePage(activePage);
    })();
  };

  const refresh = async () => {
    setLoading(true);
    await loadChannels(activePage - 1);
  };

  useEffect(() => {
    loadChannels(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []);
  // 监听频道列表变化
  useEffect(() => {
    if (channels.length > 0) {
      fetchMonthlyQuotasAndChannels(channels);
    }
  }, [channels]);
  
  


  const manageChannel = async (id, action, idx, value) => {
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/channel/${id}/`);
        break;
      case 'enable':
        data.status = 1;
        res = await API.put('/api/channel/', data);
        break;
      case 'disable':
        data.status = 2;
        res = await API.put('/api/channel/', data);
        break;
      case 'priority':
        if (value === '') {
          return;
        }
        data.priority = parseInt(value);
        res = await API.put('/api/channel/', data);
        break;
      case 'weight':
        if (value === '') {
          return;
        }
        data.weight = parseInt(value);
        if (data.weight < 0) {
          data.weight = 0;
        }
        res = await API.put('/api/channel/', data);
        break;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess('操作成功完成！');
      let channel = res.data.data;
      let newChannels = [...channels];
      let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      if (action === 'delete') {
        newChannels[realIdx].deleted = true;
      } else {
        newChannels[realIdx].status = channel.status;
      }
      setChannels(newChannels);
    } else {
      showError(message);
    }
  };

  const renderStatus = (status) => {
    switch (status) {
      case 1:
        return <Label basic style={{ color: 'var(--czl-success-color)' }}>已启用</Label>;
      case 2:
        return (
          <Popup
            trigger={<Label basic style={{ color: 'var(--czl-error-color)' }}>
              已禁用
            </Label>}
            content='本渠道被手动禁用'
            basic
          />
        );
      case 3:
        return (
          <Popup
            trigger={<Label basic style={{ color: 'var(--czl-warning-color)' }}>
              已禁用
            </Label>}
            content='本渠道被程序自动禁用'
            basic
          />
        );
      default:
        return (
          <Label basic style={{ color: 'var(--czl-grayC)' }}>
            未知状态
          </Label>
        );

    }
  };

  const renderResponseTime = (responseTime) => {
    let time = responseTime / 1000;
    time = time.toFixed(2) + ' 秒';
    if (responseTime === 0) {
      return <Label basic style={{ color: 'var(--czl-grayA)' }}>未测试</Label>;
    } else if (responseTime <= 1000) {
      return <Label basic style={{ color: 'var(--czl-success-color)' }}>{time}</Label>;
    } else if (responseTime <= 3000) {
      return <Label basic style={{ color: 'var(--czl-primary-color)' }}>{time}</Label>;
    } else if (responseTime <= 5000) {
      return <Label basic style={{ color: 'var(--czl-warning-color)' }}>{time}</Label>;
    } else {
      return <Label basic style={{ color: 'var(--czl-error-color)' }}>{time}</Label>;
    }
  };

  const searchChannels = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load files instead.
      await loadChannels(0);
      setActivePage(1);
      return;
    }
    setSearching(true);
    const res = await API.get(`/api/channel/search?keyword=${searchKeyword}`);
    const { success, message, data } = res.data;
    if (success) {
      setChannels(data);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const testChannel = async (id, name, idx) => {
    const res = await API.get(`/api/channel/test/${id}/`);
    const { success, message, time } = res.data;
    if (success) {
      let newChannels = [...channels];
      let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      newChannels[realIdx].response_time = time * 1000;
      newChannels[realIdx].test_time = Date.now() / 1000;
      setChannels(newChannels);
      showInfo(`通道 ${name} 测试成功，耗时 ${time.toFixed(2)} 秒。`);
    } else {
      showError(message);
    }
  };

  const testAllChannels = async () => {
    const res = await API.get(`/api/channel/test`);
    const { success, message } = res.data;
    if (success) {
      showInfo('已成功开始测试所有已启用通道，请刷新页面查看结果。');
    } else {
      showError(message);
    }
  };

  const deleteAllDisabledChannels = async () => {
    const res = await API.delete(`/api/channel/disabled`);
    const { success, message, data } = res.data;
    if (success) {
      showSuccess(`已删除所有禁用渠道，共计 ${data} 个`);
      await refresh();
    } else {
      showError(message);
    }
  };

  const updateChannelBalance = async (id, name, idx) => {
    const res = await API.get(`/api/channel/update_balance/${id}/`);
    const { success, message, balance } = res.data;
    if (success) {
      let newChannels = [...channels];
      let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      newChannels[realIdx].balance = balance;
      newChannels[realIdx].balance_updated_time = Date.now() / 1000;
      setChannels(newChannels);
      showInfo(`通道 ${name} 余额更新成功！`);
    } else {
      showError(message);
    }
  };

  const updateAllChannelsBalance = async () => {
    setUpdatingBalance(true);
    const res = await API.get(`/api/channel/update_balance`);
    const { success, message } = res.data;
    if (success) {
      showInfo('已更新完毕所有已启用通道余额！');
    } else {
      showError(message);
    }
    setUpdatingBalance(false);
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortChannel = (key) => {
    if (channels.length === 0) return;
    setLoading(true);
    let sortedChannels = [...channels];
    sortedChannels.sort((a, b) => {
      if (!isNaN(a[key])) {
        // If the value is numeric, subtract to sort
        return a[key] - b[key];
      } else {
        // If the value is not numeric, sort as strings
        return ('' + a[key]).localeCompare(b[key]);
      }
    });
    if (sortedChannels[0].id === channels[0].id) {
      sortedChannels.reverse();
    }
    setChannels(sortedChannels);
    setLoading(false);
  };

  // Truncate string
  function truncateString(str, num) {
    if (str.length <= num) return str;
    return str.slice(0, num) + "...";
  }
  // 总已用额度
  function formatUsedQuota(usedQuota) {
    const quotaPerUnit = localStorage.getItem('quota_per_unit') || 500000; // 如果未设置，则使用 1 作为默认值
    return `$${(usedQuota / quotaPerUnit).toFixed(3)}`;
  }






  return (
    <>
      <Form onSubmit={searchChannels}>
        <Form.Input
          icon='search'
          fluid
          iconPosition='left'
          placeholder='搜索渠道的 ID，名称和密钥 ...'
          value={searchKeyword}
          loading={searching}
          onChange={handleKeywordChange}
        />
      </Form>
      <Table basic compact size='small'>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('id');
              }}
            >
              ID
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('name');
              }}
            >
              名称
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('group');
              }}
            >
              分组
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('type');
              }}
            >
              类型
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('status');
              }}
            >
              状态
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('response_time');
              }}
            >
              响应时间
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('base_url');
              }}
            >
              Base URL
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('models');
              }}
            >
              模型
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('used_quota');
              }}
            >
              本月
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('used_quota');
              }}
            >
              总共
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('balance');
              }}
            >
              余额
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('priority');
              }}
            >
              优先级
            </Table.HeaderCell>
            <Table.HeaderCell>操作</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {channels
            .slice(
              (activePage - 1) * ITEMS_PER_PAGE,
              activePage * ITEMS_PER_PAGE
            )
            .map((channel, idx) => {
              if (channel.deleted) return <></>;
              return (
                <Table.Row key={channel.id}>
                  <Table.Cell>{channel.id}</Table.Cell>
                  <Table.Cell>{channel.name ? channel.name : '无'}</Table.Cell>
                  <Table.Cell>{renderGroup(channel.group)}</Table.Cell>
                  <Table.Cell>{renderType(channel.type)}</Table.Cell>
                  <Table.Cell>{renderStatus(channel.status)}</Table.Cell>
                  <Table.Cell>
                    <Popup
                      content={channel.test_time ? renderTimestamp(channel.test_time) : '未测试'}
                      key={channel.id}
                      trigger={renderResponseTime(channel.response_time)}
                      basic
                    />
                  </Table.Cell>


                  <Table.Cell>
                    <Popup
                      content={channel.base_url}
                      trigger={<span>{truncateString(channel.base_url.replace(/^https?:\/\//, ''), 10)}</span>}
                      basic
                    />
                  </Table.Cell>
                  <Table.Cell>
                    <Popup
                      content={channel.models}
                      trigger={<span>{truncateString(channel.models, 10)}</span>}
                      basic
                    />
                  </Table.Cell>
                  <Table.Cell>
                  <Label basic style={{ color: 'var(--czl-blue-500)',border:'1px solid var(--czl-blue-500)' }}>${monthlyQuotas[channel.id]}</Label>
                  </Table.Cell>
                  <Table.Cell>
                    <Label basic style={{ color: 'var(--czl-blue-800)',border:'1px solid var(--czl-blue-800)' }}>{formatUsedQuota(channel.used_quota)}</Label>
                  </Table.Cell>
                  <Table.Cell>
                    <Popup
                      trigger={<span onClick={() => {
                        updateChannelBalance(channel.id, channel.name, idx);
                      }} style={{ cursor: 'pointer' }}>
                        {renderBalance(channel.type, channel.balance)}
                      </span>}
                      content='点击更新'
                      basic
                    />
                  </Table.Cell>
                  <Table.Cell>
                    <Popup
                      trigger={<Input type='number' defaultValue={channel.priority} onBlur={(event) => {
                        manageChannel(
                          channel.id,
                          'priority',
                          idx,
                          event.target.value
                        );
                      }}>
                        <input style={{ maxWidth: '60px' }} />
                      </Input>}
                      content='渠道选择优先级，越高越优先'
                      basic
                    />
                  </Table.Cell>
                  <Table.Cell>
                    <div>
                      <Button
                        icon='play' // 示例图标
                        size={'small'}
                        positive
                        style={{ backgroundColor: 'var(--czl-success-color)' }}
                        onClick={() => {
                          testChannel(channel.id, channel.name, idx);
                        }}
                      />

                      <Popup
                        trigger={
                          <Button icon='trash' size='small' negative style={{ backgroundColor: 'var(--czl-error-color)' }} />
                        }
                        on='click'
                        flowing
                        hoverable
                      >
                        <Button
                          negative
                          style={{ backgroundColor: 'var(--czl-error-color)' }}
                          onClick={() => {
                            manageChannel(channel.id, 'delete', idx);
                          }}
                        >
                          确认删除 {channel.name}
                        </Button>
                      </Popup>

                      <Button
                        icon={channel.status === 1 ? 'ban' : 'check'} // 示例图标
                        size={'small'}
                        negative
                        style={{ backgroundColor: 'var(--czl-warning-color)' }}
                        onClick={() => {
                          manageChannel(
                            channel.id,
                            channel.status === 1 ? 'disable' : 'enable',
                            idx
                          );
                        }}
                      />

                      <Button
                        icon='edit' // 示例图标
                        size={'small'}
                        as={Link}
                        to={'/channel/edit/' + channel.id}
                        style={{ backgroundColor: 'var(--czl-primary-color)' }}
                      />
                    </div>
                  </Table.Cell>

                </Table.Row>
              );
            })}
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan='13'>
              <Button size='small' as={Link} to='/channel/add' loading={loading}>
                添加新的渠道
              </Button>
              <Button size='small' loading={loading} onClick={testAllChannels}>
                测试所有已启用通道
              </Button>
              <Button size='small' onClick={updateAllChannelsBalance}
                loading={loading || updatingBalance}>更新所有已启用通道余额</Button>
              <Popup
                trigger={
                  <Button size='small' loading={loading}>
                    删除禁用渠道
                  </Button>
                }
                on='click'
                flowing
                hoverable
              >
                <Button size='small' loading={loading} negative onClick={deleteAllDisabledChannels}>
                  确认删除
                </Button>
              </Popup>
              <Pagination
                floated='right'
                activePage={activePage}
                onPageChange={onPaginationChange}
                size='small'
                siblingRange={1}
                totalPages={
                  Math.ceil(channels.length / ITEMS_PER_PAGE) +
                  (channels.length % ITEMS_PER_PAGE === 0 ? 1 : 0)
                }
              />
              <Button size='small' onClick={refresh} loading={loading}>刷新</Button>
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
    </>
  );
};

export default ChannelsTable;
