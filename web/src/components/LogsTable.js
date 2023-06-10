import React, { useEffect, useState } from 'react';
import { Button, Label, Pagination, Select, Table } from 'semantic-ui-react';
import { API, isAdmin, showError, timestamp2string } from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';

function renderTimestamp(timestamp) {
  return (
    <>
      {timestamp2string(timestamp)}
    </>
  );
}

const MODE_OPTIONS = [
  { key: 'all', text: '全部用户', value: 'all' },
  { key: 'self', text: '当前用户', value: 'self' },
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
      return <Label basic color='green'> 充值 </Label>;
    case 2:
      return <Label basic color='olive'> 消费 </Label>;
    case 3:
      return <Label basic color='orange'> 管理 </Label>;
    case 4:
      return <Label basic color='purple'> 系统 </Label>;
    default:
      return <Label basic color='black'> 未知 </Label>;
  }
}

const LogsTable = () => {
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [logType, setLogType] = useState(0);
  const [mode, setMode] = useState('self'); // all, self
  const showModePanel = isAdmin();

  const loadLogs = async (startIdx) => {
    let url = `/api/log/self/?p=${startIdx}&type=${logType}`;
    if (mode === 'all') {
      url = `/api/log/?p=${startIdx}&type=${logType}`;
    }
    const res = await API.get(url);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setLogs(data);
      } else {
        let newLogs = logs;
        newLogs.push(...data);
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
    await loadLogs(0);
  };

  useEffect(() => {
    loadLogs(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []);

  useEffect(() => {
    refresh().then();
  }, [mode, logType]);

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
    sortedLogs.sort((a, b) => {
      return ('' + a[key]).localeCompare(b[key]);
    });
    if (sortedLogs[0].id === logs[0].id) {
      sortedLogs.reverse();
    }
    setLogs(sortedLogs);
    setLoading(false);
  };

  return (
    <>
      <Table basic>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortLog('created_time');
              }}
              width={3}
            >
              时间
            </Table.HeaderCell>
            {
              showModePanel && (
                <Table.HeaderCell
                  style={{ cursor: 'pointer' }}
                  onClick={() => {
                    sortLog('user_id');
                  }}
                  width={1}
                >
                  用户
                </Table.HeaderCell>
              )
            }
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortLog('type');
              }}
              width={2}
            >
              类型
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortLog('content');
              }}
              width={showModePanel ? 10 : 11}
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
                <Table.Row key={log.created_at}>
                  <Table.Cell>{renderTimestamp(log.created_at)}</Table.Cell>
                  {
                    showModePanel && (
                      <Table.Cell><Label>{log.user_id}</Label></Table.Cell>
                    )
                  }
                  <Table.Cell>{renderType(log.type)}</Table.Cell>
                  <Table.Cell>{log.content}</Table.Cell>
                </Table.Row>
              );
            })}
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan={showModePanel ? '5' : '4'}>
              {
                showModePanel && (
                  <Select
                    placeholder='选择模式'
                    options={MODE_OPTIONS}
                    style={{ marginRight: '8px' }}
                    name='mode'
                    value={mode}
                    onChange={(e, { name, value }) => {
                      setMode(value);
                    }}
                  />
                )
              }
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
              />
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
    </>
  );
};

export default LogsTable;
