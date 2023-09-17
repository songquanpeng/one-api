import React, { useEffect, useState } from 'react';
import { Button, Dropdown, Form, Label, Pagination, Popup, Table } from 'semantic-ui-react';
import { Link } from 'react-router-dom';
import { API, copy, showError, showSuccess, showWarning, timestamp2string } from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';
import { renderQuota } from '../helpers/render';

const COPY_OPTIONS = [
  { key: 'next', text: 'ChatGPT Next Web', value: 'next' },
  { key: 'ama', text: 'AMA 问天', value: 'ama' },
  { key: 'opencat', text: 'OpenCat', value: 'opencat' },
];

const OPEN_LINK_OPTIONS = [
  { key: 'ama', text: 'AMA 问天', value: 'ama' },
  { key: 'opencat', text: 'OpenCat', value: 'opencat' },
];

function renderTimestamp(timestamp) {
  return (
    <>
      {timestamp2string(timestamp)}
    </>
  );
}

function renderStatus(status) {
  switch (status) {
    case 1:
      return <Label basic color='green'>已启用</Label>;
    case 2:
      return <Label basic color='red'> 已禁用 </Label>;
    case 3:
      return <Label basic color='yellow'> 已过期 </Label>;
    case 4:
      return <Label basic color='grey'> 已耗尽 </Label>;
    default:
      return <Label basic color='black'> 未知状态 </Label>;
  }
}

const TokensTable = () => {
  const [tokens, setTokens] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [showTopUpModal, setShowTopUpModal] = useState(false);
  const [targetTokenIdx, setTargetTokenIdx] = useState(0);

  const loadTokens = async (startIdx) => {
    const res = await API.get(`/api/token/?p=${startIdx}`);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setTokens(data);
      } else {
        let newTokens = [...tokens];
        newTokens.splice(startIdx * ITEMS_PER_PAGE, data.length, ...data);
        setTokens(newTokens);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(tokens.length / ITEMS_PER_PAGE) + 1) {
        // In this case we have to load more data and then append them.
        await loadTokens(activePage - 1);
      }
      setActivePage(activePage);
    })();
  };

  const refresh = async () => {
    setLoading(true);
    await loadTokens(activePage - 1);
  };

  const onCopy = async (type, key) => {
    let status = localStorage.getItem('status');
    let serverAddress = '';
    if (status) {
      status = JSON.parse(status);
      serverAddress = status.server_address;
    }
    if (serverAddress === '') {
      serverAddress = window.location.origin;
    }
    let encodedServerAddress = encodeURIComponent(serverAddress);
    const nextLink = localStorage.getItem('chat_link');
    let nextUrl;
  
    if (nextLink) {
      nextUrl = nextLink + `/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
    } else {
      nextUrl = `https://chat.oneapi.pro/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
    }

    let url;
    switch (type) {
      case 'ama':
        url = `ama://set-api-key?server=${encodedServerAddress}&key=sk-${key}`;
        break;
      case 'opencat':
        url = `opencat://team/join?domain=${encodedServerAddress}&token=sk-${key}`;
        break;
      case 'next':
        url = nextUrl;
        break;
      default:
        url = `sk-${key}`;
    }
    if (await copy(url)) {
      showSuccess('已复制到剪贴板！');
    } else {
      showWarning('无法复制到剪贴板，请手动复制，已将令牌填入搜索框。');
      setSearchKeyword(url);
    }
  };

  const onOpenLink = async (type, key) => {
    let status = localStorage.getItem('status');
    let serverAddress = '';
    if (status) {
      status = JSON.parse(status);
      serverAddress = status.server_address; 
    }
    if (serverAddress === '') {
      serverAddress = window.location.origin;
    }
    let encodedServerAddress = encodeURIComponent(serverAddress);
    const chatLink = localStorage.getItem('chat_link');
    let defaultUrl;
  
    if (chatLink) {
      defaultUrl = chatLink + `/#/?settings={"key":"sk-${key}"}`;
    } else {
      defaultUrl = `https://chat.oneapi.pro/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
    }
    let url;
    switch (type) {
      case 'ama':
        url = `ama://set-api-key?server=${encodedServerAddress}&key=sk-${key}`;
        break;
  
      case 'opencat':
        url = `opencat://team/join?domain=${encodedServerAddress}&token=sk-${key}`;
        break;
        
      default:
        url = defaultUrl;
    }
  
    window.open(url, '_blank');
  }

  useEffect(() => {
    loadTokens(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []);

  const manageToken = async (id, action, idx) => {
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/token/${id}/`);
        break;
      case 'enable':
        data.status = 1;
        res = await API.put('/api/token/?status_only=true', data);
        break;
      case 'disable':
        data.status = 2;
        res = await API.put('/api/token/?status_only=true', data);
        break;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess('操作成功完成！');
      let token = res.data.data;
      let newTokens = [...tokens];
      let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      if (action === 'delete') {
        newTokens[realIdx].deleted = true;
      } else {
        newTokens[realIdx].status = token.status;
      }
      setTokens(newTokens);
    } else {
      showError(message);
    }
  };

  const searchTokens = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load files instead.
      await loadTokens(0);
      setActivePage(1);
      return;
    }
    setSearching(true);
    const res = await API.get(`/api/token/search?keyword=${searchKeyword}`);
    const { success, message, data } = res.data;
    if (success) {
      setTokens(data);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortToken = (key) => {
    if (tokens.length === 0) return;
    setLoading(true);
    let sortedTokens = [...tokens];
    sortedTokens.sort((a, b) => {
      return ('' + a[key]).localeCompare(b[key]);
    });
    if (sortedTokens[0].id === tokens[0].id) {
      sortedTokens.reverse();
    }
    setTokens(sortedTokens);
    setLoading(false);
  };

  return (
    <>
      <Form onSubmit={searchTokens}>
        <Form.Input
          icon='search'
          fluid
          iconPosition='left'
          placeholder='搜索令牌的名称 ...'
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
                sortToken('name');
              }}
            >
              名称
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('status');
              }}
            >
              状态
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('used_quota');
              }}
            >
              已用额度
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('remain_quota');
              }}
            >
              剩余额度
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('created_time');
              }}
            >
              创建时间
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('expired_time');
              }}
            >
              过期时间
            </Table.HeaderCell>
            <Table.HeaderCell>操作</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {tokens
            .slice(
              (activePage - 1) * ITEMS_PER_PAGE,
              activePage * ITEMS_PER_PAGE
            )
            .map((token, idx) => {
              if (token.deleted) return <></>;
              return (
                <Table.Row key={token.id}>
                  <Table.Cell>{token.name ? token.name : '无'}</Table.Cell>
                  <Table.Cell>{renderStatus(token.status)}</Table.Cell>
                  <Table.Cell>{renderQuota(token.used_quota)}</Table.Cell>
                  <Table.Cell>{token.unlimited_quota ? '无限制' : renderQuota(token.remain_quota, 2)}</Table.Cell>
                  <Table.Cell>{renderTimestamp(token.created_time)}</Table.Cell>
                  <Table.Cell>{token.expired_time === -1 ? '永不过期' : renderTimestamp(token.expired_time)}</Table.Cell>
                  <Table.Cell>
                    <div>
                    <Button.Group color='green' size={'small'}>
                        <Button
                          size={'small'}
                          positive
                          onClick={async () => {
                            await onCopy('', token.key);
                          }}
                        >
                          复制
                        </Button>
                        <Dropdown
                          className='button icon'
                          floating
                          options={COPY_OPTIONS.map(option => ({
                            ...option,
                            onClick: async () => {
                              await onCopy(option.value, token.key);
                            }
                          }))}
                          trigger={<></>}
                        />
                      </Button.Group>
                      {' '}
                      <Button.Group color='blue' size={'small'}>
                        <Button
                            size={'small'}
                            positive
                            onClick={() => {     
                              onOpenLink('', token.key);       
                            }}>
                            聊天
                          </Button>
                          <Dropdown   
                            className="button icon"       
                            floating
                            options={OPEN_LINK_OPTIONS.map(option => ({
                              ...option,
                              onClick: async () => {
                                await onOpenLink(option.value, token.key);
                              }
                            }))}       
                            trigger={<></>}   
                          />
                      </Button.Group>
                      {' '}
                      <Popup
                        trigger={
                          <Button size='small' negative>
                            删除
                          </Button>
                        }
                        on='click'
                        flowing
                        hoverable
                      >
                        <Button
                          negative
                          onClick={() => {
                            manageToken(token.id, 'delete', idx);
                          }}
                        >
                          删除令牌 {token.name}
                        </Button>
                      </Popup>
                      <Button
                        size={'small'}
                        onClick={() => {
                          manageToken(
                            token.id,
                            token.status === 1 ? 'disable' : 'enable',
                            idx
                          );
                        }}
                      >
                        {token.status === 1 ? '禁用' : '启用'}
                      </Button>
                      <Button
                        size={'small'}
                        as={Link}
                        to={'/token/edit/' + token.id}
                      >
                        编辑
                      </Button>
                    </div>
                  </Table.Cell>
                </Table.Row>
              );
            })}
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan='7'>
              <Button size='small' as={Link} to='/token/add' loading={loading}>
                添加新的令牌
              </Button>
              <Button size='small' onClick={refresh} loading={loading}>刷新</Button>
              <Pagination
                floated='right'
                activePage={activePage}
                onPageChange={onPaginationChange}
                size='small'
                siblingRange={1}
                totalPages={
                  Math.ceil(tokens.length / ITEMS_PER_PAGE) +
                  (tokens.length % ITEMS_PER_PAGE === 0 ? 1 : 0)
                }
              />
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
    </>
  );
};

export default TokensTable;
