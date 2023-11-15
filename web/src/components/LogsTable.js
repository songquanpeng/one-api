import React, { useEffect, useState } from 'react';
import { Button, Form, Header, Label, Pagination, Segment, Select, Table } from 'semantic-ui-react';
import { API, isAdmin, showError, timestamp2string } from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';
import { renderQuota } from '../helpers/render';

function renderTimestamp(timestamp) {
  return (
    <>
      {timestamp2string(timestamp)}
    </>
  );
}

const MODE_OPTIONS = [
  { key: 'all', text: '全部用户', value: 'all' },
  { key: 'self', text: '当前用户', value: 'self' }
];

const LOG_OPTIONS = [
  { key: '0', text: '全部', value: 0 },
  { key: '1', text: '充值', value: 1 },
  { key: '2', text: '消费', value: 2 },
  { key: '3', text: '管理', value: 3 },
  { key: '4', text: '系统', value: 4 }
];

function renderType(type) {
  switch (type) {
    case 1:
      return <Label basic style={{ color: 'var(--czl-success-color)' }}> 充值 </Label>;
    case 2:
      return <Label basic style={{ color: 'var(--czl-primary-color)' }}> 消费 </Label>;
    case 3:
      return <Label basic style={{ color: 'var(--czl-warning-color)' }}> 管理 </Label>;
    case 4:
      return <Label basic style={{ color: 'var(--czl-primary-color-suppl-dark)' }}> 系统 </Label>;
    default:
      return <Label basic style={{ color: 'var(--czl-primary-color-dark)' }}> 未知 </Label>;
  }
}

const LogsTable = () => {
  const [logs, setLogs] = useState([]);
  const [showStat, setShowStat] = useState(false);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [logType, setLogType] = useState(0);
  const isAdminUser = isAdmin();
  let now = new Date();
  const [inputs, setInputs] = useState({
    username: '',
    token_name: '',
    model_name: '',
    start_timestamp: timestamp2string(0),
    end_timestamp: timestamp2string(now.getTime() / 1000 + 3600),
    channel: ''
  });
  const { username, token_name, model_name, start_timestamp, end_timestamp, channel } = inputs;

  const [stat, setStat] = useState({
    quota: 0,
    token: 0
  });

  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  // 快速选时间
  const handleTimePresetClick = (preset) => {
    let start, end;

    switch (preset) {
      case 'today':
        start = new Date();
        start.setHours(0, 0, 0, 0);
        end = new Date();
        end.setHours(23, 59, 59, 999);
        break;
      case 'yesterday':
        start = new Date(Date.now() - 24 * 60 * 60 * 1000);
        start.setHours(0, 0, 0, 0);
        end = new Date(Date.now() - 24 * 60 * 60 * 1000);
        end.setHours(23, 59, 59, 999);
        break;
      case 'week':
        start = new Date();
        start.setDate(start.getDate() - start.getDay() + (start.getDay() === 0 ? -6 : 1));
        start.setHours(0, 0, 0, 0);
        end = new Date();
        end.setHours(23, 59, 59, 999);
        break;
      case 'lastWeek':
        start = new Date();
        start.setDate(start.getDate() - start.getDay() + (start.getDay() === 0 ? -13 : -6));
        start.setHours(0, 0, 0, 0);
        end = new Date(start);
        end.setDate(end.getDate() + 6);
        end.setHours(23, 59, 59, 999);
        break;
      case '30days':
        start = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);
        start.setHours(0, 0, 0, 0);
        end = new Date();
        end.setHours(23, 59, 59, 999);
        break;
      case 'month':
        start = new Date();
        start.setDate(1);
        start.setHours(0, 0, 0, 0);
        end = new Date();
        end.setHours(23, 59, 59, 999);
        break;
      case 'lastMonth':
        start = new Date();
        start.setDate(1);
        start.setMonth(start.getMonth() - 1);
        start.setHours(0, 0, 0, 0);
        end = new Date(start);
        end.setMonth(end.getMonth() + 1);
        end.setDate(end.getDate() - 1);
        end.setHours(23, 59, 59, 999);
        break;
      case 'reset':
        start = new Date(0);
        end = new Date(now.getTime() + 3600 * 1000);
        break;
      default:
        break;

    }

    setInputs((inputs) => ({
      ...inputs,
      start_timestamp: timestamp2string(start.getTime() / 1000),
      end_timestamp: timestamp2string(end.getTime() / 1000),
    }));
  };



  const getLogSelfStat = async () => {
    let localStartTimestamp = Date.parse(start_timestamp) / 1000;
    let localEndTimestamp = Date.parse(end_timestamp) / 1000;
    let res = await API.get(`/api/log/self/stat?type=${logType}&token_name=${token_name}&model_name=${model_name}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}`);
    const { success, message, data } = res.data;
    if (success) {
      setStat(data);
    } else {
      showError(message);
    }
  };

  const getLogStat = async () => {
    let localStartTimestamp = Date.parse(start_timestamp) / 1000;
    let localEndTimestamp = Date.parse(end_timestamp) / 1000;
    let res = await API.get(`/api/log/stat?type=${logType}&username=${username}&token_name=${token_name}&model_name=${model_name}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}&channel=${channel}`);
    const { success, message, data } = res.data;
    if (success) {
      setStat(data);
    } else {
      showError(message);
    }
  };

  const handleEyeClick = async () => {
    if (isAdminUser) {
      await getLogStat();
    } else {
      await getLogSelfStat();
    }
  };

  const loadLogs = async (startIdx) => {
    let url = '';
    let localStartTimestamp = Date.parse(start_timestamp) / 1000;
    let localEndTimestamp = Date.parse(end_timestamp) / 1000;
    if (isAdminUser) {
      url = `/api/log/?p=${startIdx}&type=${logType}&username=${username}&token_name=${token_name}&model_name=${model_name}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}&channel=${channel}`;
    } else {
      url = `/api/log/self/?p=${startIdx}&type=${logType}&token_name=${token_name}&model_name=${model_name}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}`;
    }
    const res = await API.get(url);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setLogs(data);
      } else {
        let newLogs = [...logs];
        newLogs.splice(startIdx * ITEMS_PER_PAGE, data.length, ...data);
        setLogs(newLogs);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(logs.length / ITEMS_PER_PAGE) + 1) {
        // In this case we have to load more data and then append them.
        await loadLogs(activePage - 1);
      }
      setActivePage(activePage);
    })();
  };

  const refresh = async () => {
    setLoading(true);
    setActivePage(1);
    await loadLogs(0);
    handleEyeClick();
  };

  useEffect(() => {
    refresh().then();
  }, [logType, username, token_name, model_name, channel, start_timestamp, end_timestamp]);

  const searchLogs = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load files instead.
      await loadLogs(0);
      setActivePage(1);
      return;
    }
    setSearching(true);
    const res = await API.get(`/api/log/self/search?keyword=${searchKeyword}`);
    const { success, message, data } = res.data;
    if (success) {
      setLogs(data);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortLog = (key) => {
    if (logs.length === 0) return;
    setLoading(true);
    let sortedLogs = [...logs];
    if (typeof sortedLogs[0][key] === 'string') {
      sortedLogs.sort((a, b) => {
        return ('' + a[key]).localeCompare(b[key]);
      });
    } else {
      sortedLogs.sort((a, b) => {
        if (a[key] === b[key]) return 0;
        if (a[key] > b[key]) return -1;
        if (a[key] < b[key]) return 1;
      });
    }
    if (sortedLogs[0].id === logs[0].id) {
      sortedLogs.reverse();
    }
    setLogs(sortedLogs);
    setLoading(false);
  };

  return (
    <>
      <Segment>
        <Header as='h3'>
          使用明细【消耗额度：
          <span style={{ color: 'var(--czl-primary-color)' }}>{renderQuota(stat.quota)}</span>】
        </Header>

        <Form>
          <Form.Group>
            <Form.Input fluid label={'Key名称'} width={3} value={token_name}
              placeholder={'可选值'} name='token_name' onChange={handleInputChange} />
            <Form.Input fluid label='模型名称' width={3} value={model_name} placeholder='可选值'
              name='model_name'
              onChange={handleInputChange} />
            <Form.Input fluid label='起始时间' width={4} value={start_timestamp} type='datetime-local'
              name='start_timestamp'
              onChange={handleInputChange} />
            <Form.Input fluid label='结束时间' width={4} value={end_timestamp} type='datetime-local'
              name='end_timestamp'
              onChange={handleInputChange} />
            <Form.Button fluid label='操作' width={2} onClick={refresh}>查询</Form.Button>


          </Form.Group>
          <Form.Group>
            {isAdminUser && (
              <>
                <Form.Input fluid label={'渠道 ID'} width={3} value={channel}
                  placeholder='可选值' name='channel'
                  onChange={handleInputChange} />
                <Form.Input fluid label={'用户名称'} width={3} value={username}
                  placeholder={'可选值'} name='username'
                  onChange={handleInputChange} />
              </>
            )}
            {/* 将按钮组移动到这里 */}
            <Form.Field>
              <label>筛选时间</label>
              <Button.Group>
                <Button onClick={() => handleTimePresetClick('today')}>今天</Button>
                <Button onClick={() => handleTimePresetClick('yesterday')}>昨天</Button>
                <Button onClick={() => handleTimePresetClick('week')}>本周</Button>
                <Button onClick={() => handleTimePresetClick('lastWeek')}>上周</Button>
                <Button onClick={() => handleTimePresetClick('month')}>本月</Button>
                <Button onClick={() => handleTimePresetClick('lastMonth')}>上月</Button>
                <Button onClick={() => handleTimePresetClick('30days')}>30天内</Button>
                <Button onClick={() => handleTimePresetClick('reset')}>重置</Button>

              </Button.Group>
            </Form.Field>
          </Form.Group>
        </Form>
        <Table basic compact size='small'>
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  sortLog('created_time');
                }}
                width={2}
              >
                时间
              </Table.HeaderCell>
              {
                isAdminUser && <Table.HeaderCell
                  style={{ cursor: 'pointer' }}
                  onClick={() => {
                    sortLog('channel');
                  }}
                  width={1}
                >
                  渠道
                </Table.HeaderCell>
              }
              {
                isAdminUser && <Table.HeaderCell
                  style={{ cursor: 'pointer' }}
                  onClick={() => {
                    sortLog('username');
                  }}
                  width={1}
                >
                  用户
                </Table.HeaderCell>
              }
              <Table.HeaderCell
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  sortLog('token_name');
                }}
                width={1}
              >
                Key
              </Table.HeaderCell>
              <Table.HeaderCell
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  sortLog('type');
                }}
                width={1}
              >
                类型
              </Table.HeaderCell>
              <Table.HeaderCell
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  sortLog('model_name');
                }}
                width={3}
              >
                模型
              </Table.HeaderCell>
              <Table.HeaderCell
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  sortLog('prompt_tokens');
                }}
                width={1}
              >
                输入
              </Table.HeaderCell>
              <Table.HeaderCell
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  sortLog('completion_tokens');
                }}
                width={1}
              >
                输出
              </Table.HeaderCell>
              <Table.HeaderCell
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  sortLog('quota');
                }}
                width={1}
              >
                额度
              </Table.HeaderCell>
              <Table.HeaderCell
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  sortLog('content');
                }}
                width={isAdminUser ? 4 : 6}
              >
                详情
              </Table.HeaderCell>
            </Table.Row>
          </Table.Header>

          <Table.Body>
            {logs
              .slice(
                (activePage - 1) * ITEMS_PER_PAGE,
                activePage * ITEMS_PER_PAGE
              )
              .map((log, idx) => {
                if (log.deleted) return <></>;
                return (
                  <Table.Row key={log.id}>
                    <Table.Cell>{renderTimestamp(log.created_at)}</Table.Cell>
                    {
                      isAdminUser && (
                        <Table.Cell>{log.channel ? <Label basic>{log.channel}</Label> : ''}</Table.Cell>
                      )
                    }
                    {
                      isAdminUser && (
                        <Table.Cell>{log.username ? <Label>{log.username}</Label> : ''}</Table.Cell>
                      )
                    }
                    <Table.Cell>{log.token_name ? <Label basic>{log.token_name}</Label> : ''}</Table.Cell>
                    <Table.Cell>{renderType(log.type)}</Table.Cell>
                    <Table.Cell>{log.model_name ? <Label basic>{log.model_name}</Label> : ''}</Table.Cell>
                    <Table.Cell>{log.prompt_tokens ? log.prompt_tokens : ''}</Table.Cell>
                    <Table.Cell>{log.completion_tokens ? log.completion_tokens : ''}</Table.Cell>
                    <Table.Cell>{log.quota ? renderQuota(log.quota, 6) : ''}</Table.Cell>
                    <Table.Cell>{log.content}</Table.Cell>
                  </Table.Row>
                );
              })}
          </Table.Body>

          <Table.Footer>
            <Table.Row>
              <Table.HeaderCell colSpan={'10'}>
                <Select
                  placeholder='选择明细分类'
                  options={LOG_OPTIONS}
                  style={{ marginRight: '8px' }}
                  name='logType'
                  value={logType}
                  onChange={(e, { name, value }) => {
                    setLogType(value);
                  }}
                />
                <Button size='small' onClick={refresh} loading={loading}>刷新</Button>
                <Pagination
                  floated='right'
                  activePage={activePage}
                  onPageChange={onPaginationChange}
                  size='small'
                  siblingRange={1}
                  totalPages={
                    Math.ceil(logs.length / ITEMS_PER_PAGE) +
                    (logs.length % ITEMS_PER_PAGE === 0 ? 1 : 0)
                  }
                  firstItem={null}  // 不显示第一页按钮
                  lastItem={null}  // 不显示最后一页按钮
                />
              </Table.HeaderCell>
            </Table.Row>
          </Table.Footer>
        </Table>
      </Segment>
    </>
  );
};

export default LogsTable;
