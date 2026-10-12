import React, {useEffect, useRef, useState} from 'react';
import {useTranslation} from 'react-i18next';
import {
  Button,
  Checkbox,
  Dropdown,
  Form,
  Input,
  Label,
  Loader,
  Message,
  Pagination,
  Popup,
  Table,
} from 'semantic-ui-react';
import {Link} from 'react-router-dom';
import {
  API,
  loadChannelModels,
  setPromptShown,
  shouldShowPrompt,
  showError,
  showInfo,
  showSuccess,
  timestamp2string,
} from '../helpers';

import {CHANNEL_OPTIONS, ITEMS_PER_PAGE} from '../constants';
import {renderGroup, renderNumber} from '../helpers/render';

function renderTimestamp(timestamp) {
  return <>{timestamp2string(timestamp)}</>;
}

let type2label = undefined;

function renderType(type, t) {
  if (!type2label) {
    type2label = new Map();
    for (let i = 0; i < CHANNEL_OPTIONS.length; i++) {
      type2label[CHANNEL_OPTIONS[i].value] = CHANNEL_OPTIONS[i];
    }
    type2label[0] = {
      value: 0,
      text: t('channel.table.status_unknown'),
      color: 'grey',
    };
  }
  return (
    <Label basic color={type2label[type]?.color}>
      {type2label[type] ? type2label[type].text : type}
    </Label>
  );
}

function renderBalance(type, balance, t) {
  switch (type) {
    case 1: // OpenAI
        if (balance === 0) {
            return <span>{t('channel.table.balance_not_supported')}</span>;
        }
      return <span>${balance.toFixed(2)}</span>;
    case 4: // CloseAI
      return <span>¥{balance.toFixed(2)}</span>;
    case 8: // 自定义
      return <span>${balance.toFixed(2)}</span>;
    case 5: // OpenAI-SB
      return <span>¥{(balance / 10000).toFixed(2)}</span>;
    case 10: // AI Proxy
      return <span>{renderNumber(balance)}</span>;
    case 12: // API2GPT
      return <span>¥{balance.toFixed(2)}</span>;
    case 13: // AIGC2D
      return <span>{renderNumber(balance)}</span>;
    case 20: // OpenRouter
      return <span>${balance.toFixed(2)}</span>;
    case 36: // DeepSeek
      return <span>¥{balance.toFixed(2)}</span>;
    case 44: // SiliconFlow
      return <span>¥{balance.toFixed(2)}</span>;
    default:
      return <span>{t('channel.table.balance_not_supported')}</span>;
  }
}

function isShowDetail() {
  return localStorage.getItem('show_detail') === 'true';
}

const promptID = 'detail';

const ChannelsTable = () => {
  const { t } = useTranslation();
  const [channels, setChannels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [updatingBalance, setUpdatingBalance] = useState(false);
  const [showPrompt, setShowPrompt] = useState(shouldShowPrompt(promptID));
  const [showDetail, setShowDetail] = useState(isShowDetail());
  const [selectedIds, setSelectedIds] = useState(new Set());
  const [batchTesting, setBatchTesting] = useState(false);
  const [batchTestStatus, setBatchTestStatus] = useState(null);
  const [deleting, setDeleting] = useState(false);
  const batchTestTimer = useRef(null);

  const processChannelData = (channel) => {
    if (channel.models === '') {
      channel.models = [];
      channel.test_model = '';
    } else {
      channel.models = channel.models.split(',');
      if (channel.models.length > 0) {
        channel.test_model = channel.models[0];
      }
      channel.model_options = channel.models.map((model) => {
        return {
          key: model,
          text: model,
          value: model,
        };
      });
      console.log('channel', channel);
    }
    return channel;
  };

  const loadChannels = async (startIdx) => {
    const res = await API.get(`/api/channel/?p=${startIdx}`);
    const { success, message, data } = res.data;
    if (success) {
      let localChannels = data.map(processChannelData);
      if (startIdx === 0) {
        setChannels(localChannels);
      } else {
        let newChannels = [...channels];
        newChannels.splice(
          startIdx * ITEMS_PER_PAGE,
          data.length,
          ...localChannels
        );
        setChannels(newChannels);
      }
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

  const toggleShowDetail = () => {
    setShowDetail(!showDetail);
    localStorage.setItem('show_detail', (!showDetail).toString());
  };

  useEffect(() => {
    loadChannels(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
    loadChannelModels().then();
  }, []);

  useEffect(() => {
    return () => {
      if (batchTestTimer.current) {
        clearTimeout(batchTestTimer.current);
      }
    };
  }, []);

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
      showSuccess(t('channel.messages.operation_success'));
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

  const renderStatus = (status, t) => {
    switch (status) {
      case 1:
        return (
          <Label basic color='green'>
            {t('channel.table.status_enabled')}
          </Label>
        );
      case 2:
        return (
          <Popup
            trigger={
              <Label basic color='red'>
                {t('channel.table.status_disabled')}
              </Label>
            }
            content={t('channel.table.status_disabled_tip')}
            basic
          />
        );
      case 3:
        return (
          <Popup
            trigger={
              <Label basic color='yellow'>
                {t('channel.table.status_auto_disabled')}
              </Label>
            }
            content={t('channel.table.status_auto_disabled_tip')}
            basic
          />
        );
      default:
        return (
          <Label basic color='grey'>
            {t('channel.table.status_unknown')}
          </Label>
        );
    }
  };

  const renderResponseTime = (responseTime, t) => {
    let time = responseTime / 1000;
    time = time.toFixed(2) + 's';
    if (responseTime === 0) {
      return (
        <Label basic color='grey'>
          {t('channel.table.not_tested')}
        </Label>
      );
    } else if (responseTime <= 1000) {
      return (
        <Label basic color='green'>
          {time}
        </Label>
      );
    } else if (responseTime <= 3000) {
      return (
        <Label basic color='olive'>
          {time}
        </Label>
      );
    } else if (responseTime <= 5000) {
      return (
        <Label basic color='yellow'>
          {time}
        </Label>
      );
    } else {
      return (
        <Label basic color='red'>
          {time}
        </Label>
      );
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
      let localChannels = data.map(processChannelData);
      setChannels(localChannels);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const switchTestModel = async (idx, model) => {
    let newChannels = [...channels];
    let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
    newChannels[realIdx].test_model = model;
    setChannels(newChannels);
  };

  const testChannel = async (id, name, idx, m) => {
    const res = await API.get(`/api/channel/test/${id}?model=${m}`);
    const { success, message, time, model } = res.data;
    if (success) {
      let newChannels = [...channels];
      let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      newChannels[realIdx].response_time = time * 1000;
      newChannels[realIdx].test_time = Date.now() / 1000;
      setChannels(newChannels);
      showSuccess(
        t('channel.messages.test_success', { name, model, time, message })
      );
    } else {
      showError(message);
    }
    let newChannels = [...channels];
    let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
    newChannels[realIdx].response_time = time * 1000;
    newChannels[realIdx].test_time = Date.now() / 1000;
    setChannels(newChannels);
  };

  const onBatchTestFinished = (status) => {
    const results = status?.results || [];
    const failed = results.filter((r) => !r.success).map((r) => r.id);
    const passed = results.length - failed.length;
    if (failed.length > 0) {
      // 自动勾选测试不通过的渠道，便于一键批量删除
      setSelectedIds((prev) => {
        const next = new Set(prev);
        failed.forEach((id) => next.add(id));
        return next;
      });
      showInfo(
        t('channel.messages.batch_test_finished', {
          passed,
          failed: failed.length,
        })
      );
    } else if (results.length > 0) {
      showSuccess(
        t('channel.messages.batch_test_all_passed', { passed })
      );
    }
    refresh();
  };

  const startBatchTest = async (scope) => {
    if (batchTesting) return;
    setBatchTesting(true);
    setBatchTestStatus(null);
    setSelectedIds(new Set());
    const res = await API.get(`/api/channel/test?scope=${scope}`);
    const { success, message } = res.data;
    if (!success) {
      showError(message);
      setBatchTesting(false);
      return;
    }
    showInfo(t('channel.messages.test_all_started'));
    const poll = async () => {
      try {
        const statusRes = await API.get('/api/channel/batch_test/status');
        const status = statusRes.data?.data;
        if (status) {
          setBatchTestStatus(status);
          if (status.running) {
            batchTestTimer.current = setTimeout(poll, 2000);
            return;
          }
          setBatchTesting(false);
          onBatchTestFinished(status);
          return;
        }
      } catch (error) {
        // 网络异常时继续轮询
      }
      batchTestTimer.current = setTimeout(poll, 2000);
    };
    poll();
  };

  const toggleSelect = (id) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const toggleSelectAll = () => {
    const visibleIds = channels
      .slice(
        (activePage - 1) * ITEMS_PER_PAGE,
        activePage * ITEMS_PER_PAGE
      )
      .filter((channel) => !channel.deleted)
      .map((channel) => channel.id);
    const allSelected = visibleIds.every((id) => selectedIds.has(id));
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (allSelected) {
        visibleIds.forEach((id) => next.delete(id));
      } else {
        visibleIds.forEach((id) => next.add(id));
      }
      return next;
    });
  };

  const deleteSelectedChannels = async () => {
    if (selectedIds.size === 0) return;
    setDeleting(true);
    try {
      const res = await API.post('/api/channel/batch_delete', {
        ids: Array.from(selectedIds),
      });
      const { success, message, data } = res.data;
      if (success) {
        showSuccess(
          t('channel.messages.batch_delete_success', { count: data })
        );
        setSelectedIds(new Set());
        setBatchTestStatus(null);
        await refresh();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setDeleting(false);
  };

  const deleteAllDisabledChannels = async () => {
    const res = await API.delete(`/api/channel/disabled`);
    const { success, message, data } = res.data;
    if (success) {
      showSuccess(
        t('channel.messages.delete_disabled_success', { count: data })
      );
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
      showSuccess(t('channel.messages.balance_update_success', { name }));
    } else {
      showError(message);
    }
  };

  const updateAllChannelsBalance = async () => {
    setUpdatingBalance(true);
    const res = await API.get(`/api/channel/update_balance`);
    const { success, message } = res.data;
    if (success) {
      showInfo(t('channel.messages.all_balance_updated'));
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

  return (
    <>
      <Form onSubmit={searchChannels}>
        <Form.Input
          icon='search'
          fluid
          iconPosition='left'
          placeholder={t('channel.search')}
          value={searchKeyword}
          loading={searching}
          onChange={handleKeywordChange}
        />
      </Form>
      {showPrompt && (
        <Message
          onDismiss={() => {
            setShowPrompt(false);
            setPromptShown(promptID);
          }}
        >
          {t('channel.balance_notice')}
          <br />
          {t('channel.test_notice')}
          <br />
          {t('channel.detail_notice')}
        </Message>
      )}
      <Table basic={'very'} compact size='small'>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>
              <Checkbox
                checked={
                  channels
                    .slice(
                      (activePage - 1) * ITEMS_PER_PAGE,
                      activePage * ITEMS_PER_PAGE
                    )
                    .filter((channel) => !channel.deleted)
                    .length > 0 &&
                  channels
                    .slice(
                      (activePage - 1) * ITEMS_PER_PAGE,
                      activePage * ITEMS_PER_PAGE
                    )
                    .filter((channel) => !channel.deleted)
                    .every((channel) => selectedIds.has(channel.id))
                }
                onChange={toggleSelectAll}
              />
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('id');
              }}
            >
              {t('channel.table.id')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('name');
              }}
            >
              {t('channel.table.name')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('group');
              }}
            >
              {t('channel.table.group')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('type');
              }}
            >
              {t('channel.table.type')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('status');
              }}
            >
              {t('channel.table.status')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('response_time');
              }}
            >
              {t('channel.table.response_time')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('balance');
              }}
            >
              {t('channel.table.balance')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('priority');
              }}
              hidden={!showDetail}
            >
              {t('channel.table.priority')}
            </Table.HeaderCell>
            <Table.HeaderCell hidden={!showDetail}>
              {t('channel.table.test_model')}
            </Table.HeaderCell>
            <Table.HeaderCell>{t('channel.table.actions')}</Table.HeaderCell>
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
                  <Table.Cell>
                    <Checkbox
                      checked={selectedIds.has(channel.id)}
                      onChange={() => toggleSelect(channel.id)}
                    />
                  </Table.Cell>
                  <Table.Cell>{channel.id}</Table.Cell>
                  <Table.Cell>
                    {channel.name ? channel.name : t('channel.table.no_name')}
                  </Table.Cell>
                  <Table.Cell>{renderGroup(channel.group)}</Table.Cell>
                  <Table.Cell>{renderType(channel.type, t)}</Table.Cell>
                  <Table.Cell>{renderStatus(channel.status, t)}</Table.Cell>
                  <Table.Cell>
                    <Popup
                      content={
                        channel.test_time
                          ? renderTimestamp(channel.test_time)
                          : t('channel.table.not_tested')
                      }
                      key={channel.id}
                      trigger={renderResponseTime(channel.response_time, t)}
                      basic
                    />
                  </Table.Cell>
                  <Table.Cell>
                    <Popup
                      trigger={
                        <span
                          onClick={() => {
                            updateChannelBalance(channel.id, channel.name, idx);
                          }}
                          style={{ cursor: 'pointer' }}
                        >
                          {renderBalance(channel.type, channel.balance, t)}
                        </span>
                      }
                      content={t('channel.table.click_to_update')}
                      basic
                    />
                  </Table.Cell>
                  <Table.Cell hidden={!showDetail}>
                    <Popup
                      trigger={
                        <Input
                          type='number'
                          defaultValue={channel.priority}
                          onBlur={(event) => {
                            manageChannel(
                              channel.id,
                              'priority',
                              idx,
                              event.target.value
                            );
                          }}
                        >
                          <input style={{ maxWidth: '60px' }} />
                        </Input>
                      }
                      content={t('channel.table.priority_tip')}
                      basic
                    />
                  </Table.Cell>
                  <Table.Cell hidden={!showDetail}>
                    <Dropdown
                      placeholder={t('channel.table.select_test_model')}
                      selection
                      options={channel.model_options}
                      defaultValue={channel.test_model}
                      onChange={(event, data) => {
                        switchTestModel(idx, data.value);
                      }}
                    />
                  </Table.Cell>
                  <Table.Cell>
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        flexWrap: 'wrap',
                        gap: '2px',
                        rowGap: '6px',
                      }}
                    >
                      <Button
                        size={'tiny'}
                        positive
                        onClick={() => {
                          testChannel(
                            channel.id,
                            channel.name,
                            idx,
                            channel.test_model
                          );
                        }}
                      >
                        {t('channel.buttons.test')}
                      </Button>
                      <Popup
                        trigger={
                          <Button size='tiny' negative>
                            {t('channel.buttons.delete')}
                          </Button>
                        }
                        on='click'
                        flowing
                        hoverable
                      >
                        <Button
                          size={'tiny'}
                          negative
                          onClick={() => {
                            manageChannel(channel.id, 'delete', idx);
                          }}
                        >
                          {t('channel.buttons.confirm_delete')} {channel.name}
                        </Button>
                      </Popup>
                      <Button
                        size={'tiny'}
                        onClick={() => {
                          manageChannel(
                            channel.id,
                            channel.status === 1 ? 'disable' : 'enable',
                            idx
                          );
                        }}
                      >
                        {channel.status === 1
                          ? t('channel.buttons.disable')
                          : t('channel.buttons.enable')}
                      </Button>
                      <Button
                        size={'tiny'}
                        as={Link}
                        to={'/channel/edit/' + channel.id}
                      >
                        {t('channel.buttons.edit')}
                      </Button>
                    </div>
                  </Table.Cell>
                </Table.Row>
              );
            })}
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan={showDetail ? '11' : '9'}>
              <Button size='tiny' as={Link} to='/channel/add' loading={loading}>
                {t('channel.buttons.add')}
              </Button>
              <Button
                size='tiny'
                loading={loading || batchTesting}
                onClick={() => {
                  startBatchTest('all');
                }}
              >
                {t('channel.buttons.test_all')}
              </Button>
              <Button
                size='tiny'
                loading={loading || batchTesting}
                onClick={() => {
                  startBatchTest('disabled');
                }}
              >
                {t('channel.buttons.test_disabled')}
              </Button>
              <Popup
                trigger={
                  <Button size='tiny' loading={loading}>
                    {t('channel.buttons.delete_disabled')}
                  </Button>
                }
                on='click'
                flowing
                hoverable
              >
                <Button
                  size='tiny'
                  loading={loading}
                  negative
                  onClick={deleteAllDisabledChannels}
                >
                  {t('channel.buttons.confirm_delete_disabled')}
                </Button>
              </Popup>
              <Popup
                trigger={
                  <Button
                    size='tiny'
                    negative
                    loading={deleting}
                    disabled={selectedIds.size === 0}
                  >
                    {t('channel.buttons.delete_selected')} ({selectedIds.size})
                  </Button>
                }
                on='click'
                flowing
                hoverable
              >
                <Button
                  size='tiny'
                  loading={deleting}
                  negative
                  disabled={selectedIds.size === 0}
                  onClick={deleteSelectedChannels}
                >
                  {t('channel.buttons.confirm_delete_selected')}{' '}
                  {selectedIds.size}
                </Button>
              </Popup>
              {batchTesting && batchTestStatus && (
                <span style={{ marginLeft: '8px', color: 'grey' }}>
                  <Loader active inline size='tiny' />
                  {t('channel.messages.batch_testing', {
                    done: batchTestStatus.done,
                    total: batchTestStatus.total,
                  })}
                </span>
              )}
              {!batchTesting &&
                batchTestStatus &&
                batchTestStatus.results.length > 0 && (
                  <span style={{ marginLeft: '8px', color: 'grey' }}>
                    {t('channel.messages.batch_test_result_summary', {
                      passed: batchTestStatus.results.filter(
                        (r) => r.success
                      ).length,
                      failed: batchTestStatus.results.filter(
                        (r) => !r.success
                      ).length,
                    })}
                  </span>
                )}
              <Pagination
                floated='right'
                activePage={activePage}
                onPageChange={onPaginationChange}
                size='tiny'
                siblingRange={1}
                totalPages={
                  Math.ceil(channels.length / ITEMS_PER_PAGE) +
                  (channels.length % ITEMS_PER_PAGE === 0 ? 1 : 0)
                }
              />
              <Button size='tiny' onClick={refresh} loading={loading}>
                {t('channel.buttons.refresh')}
              </Button>
              <Button size='tiny' onClick={toggleShowDetail}>
                {showDetail
                  ? t('channel.buttons.hide_detail')
                  : t('channel.buttons.show_detail')}
              </Button>
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
    </>
  );
};

export default ChannelsTable;
