import React, { useEffect, useState } from 'react';
import {
  Button,
  Form,
  Label,
  Pagination,
  Popup,
  Table,
  Dropdown,
} from 'semantic-ui-react';
import { Link } from 'react-router-dom';
import { API, showError, showSuccess } from '../helpers';
import { useTranslation } from 'react-i18next';

import { ITEMS_PER_PAGE } from '../constants';
import {
  renderGroup,
  renderNumber,
  renderQuota,
  renderText,
} from '../helpers/render';

function renderRole(role, t) {
  switch (role) {
    case 1:
      return <Label>{t('user.table.role_types.normal')}</Label>;
    case 10:
      return <Label color='yellow'>{t('user.table.role_types.admin')}</Label>;
    case 100:
      return (
        <Label color='orange'>{t('user.table.role_types.super_admin')}</Label>
      );
    default:
      return <Label color='red'>{t('user.table.role_types.unknown')}</Label>;
  }
}

const UsersTable = () => {
  const { t } = useTranslation();
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [orderBy, setOrderBy] = useState('');

  const loadUsers = async (startIdx) => {
    const res = await API.get(`/api/user/?p=${startIdx}&order=${orderBy}`);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setUsers(data);
      } else {
        let newUsers = users;
        newUsers.push(...data);
        setUsers(newUsers);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(users.length / ITEMS_PER_PAGE) + 1) {
        // In this case we have to load more data and then append them.
        await loadUsers(activePage - 1, orderBy);
      }
      setActivePage(activePage);
    })();
  };

  useEffect(() => {
    loadUsers(0, orderBy)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, [orderBy]);

  const manageUser = (username, action, idx) => {
    (async () => {
      const res = await API.post('/api/user/manage', {
        username,
        action,
      });
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('user.messages.operation_success'));
        let user = res.data.data;
        let newUsers = [...users];
        let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
        if (action === 'delete') {
          newUsers[realIdx].deleted = true;
        } else {
          newUsers[realIdx].status = user.status;
          newUsers[realIdx].role = user.role;
        }
        setUsers(newUsers);
      } else {
        showError(message);
      }
    })();
  };

  const renderStatus = (status) => {
    switch (status) {
      case 1:
        return <Label basic>{t('user.table.status_types.activated')}</Label>;
      case 2:
        return (
          <Label basic color='red'>
            {t('user.table.status_types.banned')}
          </Label>
        );
      default:
        return (
          <Label basic color='grey'>
            {t('user.table.status_types.unknown')}
          </Label>
        );
    }
  };

  const searchUsers = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load files instead.
      await loadUsers(0);
      setActivePage(1);
      setOrderBy('');
      return;
    }
    setSearching(true);
    const res = await API.get(`/api/user/search?keyword=${searchKeyword}`);
    const { success, message, data } = res.data;
    if (success) {
      setUsers(data);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortUser = (key) => {
    if (users.length === 0) return;
    setLoading(true);
    let sortedUsers = [...users];
    sortedUsers.sort((a, b) => {
      if (!isNaN(a[key])) {
        // If the value is numeric, subtract to sort
        return a[key] - b[key];
      } else {
        // If the value is not numeric, sort as strings
        return ('' + a[key]).localeCompare(b[key]);
      }
    });
    if (sortedUsers[0].id === users[0].id) {
      sortedUsers.reverse();
    }
    setUsers(sortedUsers);
    setLoading(false);
  };

  const handleOrderByChange = (e, { value }) => {
    setOrderBy(value);
    setActivePage(1);
  };

  return (
    <>
      <Form onSubmit={searchUsers}>
        <Form.Input
          icon='search'
          fluid
          iconPosition='left'
          placeholder={t('user.search')}
          value={searchKeyword}
          loading={searching}
          onChange={handleKeywordChange}
        />
      </Form>

      <Table basic={'very'} compact size='small'>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('id');
              }}
            >
              {t('user.table.id')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('username');
              }}
            >
              {t('user.table.username')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('group');
              }}
            >
              {t('user.table.group')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('quota');
              }}
            >
              {t('user.table.quota')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('role');
              }}
            >
              {t('user.table.role_text')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('status');
              }}
            >
              {t('user.table.status_text')}
            </Table.HeaderCell>
            <Table.HeaderCell>{t('user.table.actions')}</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {users
            .slice(
              (activePage - 1) * ITEMS_PER_PAGE,
              activePage * ITEMS_PER_PAGE
            )
            .map((user, idx) => {
              if (user.deleted) return <></>;
              return (
                <Table.Row key={user.id}>
                  <Table.Cell>{user.id}</Table.Cell>
                  <Table.Cell>
                    <Popup
                      content={user.email ? user.email : '未绑定邮箱地址'}
                      key={user.username}
                      header={
                        user.display_name ? user.display_name : user.username
                      }
                      trigger={<span>{renderText(user.username, 15)}</span>}
                      hoverable
                    />
                  </Table.Cell>
                  <Table.Cell>{renderGroup(user.group)}</Table.Cell>
                  {/*<Table.Cell>*/}
                  {/*  {user.email ? <Popup hoverable content={user.email} trigger={<span>{renderText(user.email, 24)}</span>} /> : '无'}*/}
                  {/*</Table.Cell>*/}
                  <Table.Cell>
                    <Popup
                      content={t('user.table.remaining_quota')}
                      trigger={
                        <Label basic>{renderQuota(user.quota, t)}</Label>
                      }
                    />
                    <Popup
                      content={t('user.table.used_quota')}
                      trigger={
                        <Label basic>{renderQuota(user.used_quota, t)}</Label>
                      }
                    />
                    <Popup
                      content={t('user.table.request_count')}
                      trigger={
                        <Label basic>{renderNumber(user.request_count)}</Label>
                      }
                    />
                  </Table.Cell>
                  <Table.Cell>{renderRole(user.role, t)}</Table.Cell>
                  <Table.Cell>{renderStatus(user.status)}</Table.Cell>
                  <Table.Cell>
                    <div>
                      <Button
                        size={'small'}
                        positive
                        onClick={() => {
                          manageUser(user.username, 'promote', idx);
                        }}
                        disabled={user.role === 100}
                      >
                        {t('user.buttons.promote')}
                      </Button>
                      <Button
                        size={'small'}
                        color={'yellow'}
                        onClick={() => {
                          manageUser(user.username, 'demote', idx);
                        }}
                        disabled={user.role === 100}
                      >
                        {t('user.buttons.demote')}
                      </Button>
                      <Popup
                        trigger={
                          <Button
                            size='small'
                            negative
                            disabled={user.role === 100}
                          >
                            {t('user.buttons.delete')}
                          </Button>
                        }
                        on='click'
                        flowing
                        hoverable
                      >
                        <Button
                          negative
                          onClick={() => {
                            manageUser(user.username, 'delete', idx);
                          }}
                        >
                          {t('user.buttons.delete_user')} {user.username}
                        </Button>
                      </Popup>
                      <Button
                        size={'small'}
                        onClick={() => {
                          manageUser(
                            user.username,
                            user.status === 1 ? 'disable' : 'enable',
                            idx
                          );
                        }}
                        disabled={user.role === 100}
                      >
                        {user.status === 1
                          ? t('user.buttons.disable')
                          : t('user.buttons.enable')}
                      </Button>
                      <Button
                        size={'small'}
                        as={Link}
                        to={'/user/edit/' + user.id}
                      >
                        {t('user.buttons.edit')}
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
              <Button size='small' as={Link} to='/user/add' loading={loading}>
                {t('user.buttons.add')}
              </Button>
              <Dropdown
                placeholder={t('user.table.sort_by')}
                selection
                options={[
                  { key: '', text: t('user.table.sort.default'), value: '' },
                  {
                    key: 'quota',
                    text: t('user.table.sort.by_quota'),
                    value: 'quota',
                  },
                  {
                    key: 'used_quota',
                    text: t('user.table.sort.by_used_quota'),
                    value: 'used_quota',
                  },
                  {
                    key: 'request_count',
                    text: t('user.table.sort.by_request_count'),
                    value: 'request_count',
                  },
                ]}
                value={orderBy}
                onChange={handleOrderByChange}
                style={{ marginLeft: '10px' }}
              />
              <Pagination
                floated='right'
                activePage={activePage}
                onPageChange={onPaginationChange}
                size='small'
                siblingRange={1}
                totalPages={
                  Math.ceil(users.length / ITEMS_PER_PAGE) +
                  (users.length % ITEMS_PER_PAGE === 0 ? 1 : 0)
                }
              />
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
    </>
  );
};

export default UsersTable;
