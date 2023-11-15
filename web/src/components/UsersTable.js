import React, { useEffect, useState } from 'react';
import { Button, Form, Label, Pagination, Popup, Table, Dropdown } from 'semantic-ui-react';
import { Link } from 'react-router-dom';
import { API, showError, showSuccess } from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';
import { renderGroup, renderNumber, renderQuota, renderText } from '../helpers/render';

function renderRole(role) {
  switch (role) {
    case 1:
      return <Label>普通用户</Label>;
    case 10:
      return <Label color='var(--czl-warning-color)'>管理员</Label>;
    case 100:
      return <Label color='var(--czl-error-color)'>超级管理员</Label>;
    default:
      return <Label color='var(--czl-error-color)'>未知身份</Label>;
  }
}

const UsersTable = () => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);

  const loadUsers = async (startIdx) => {
    const res = await API.get(`/api/user/?p=${startIdx}`);
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
        await loadUsers(activePage - 1);
      }
      setActivePage(activePage);
    })();
  };

  useEffect(() => {
    loadUsers(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []);

  const manageUser = (username, action, idx, value = null) => {
    (async () => {
      let dataToSend = {
        username,
        action
      };

      // 如果是修改用户组，需要在请求体中包含新的组别信息
      if (action === 'changeGroup' && value !== null) {
        dataToSend.newGroup = value;
      }

      const res = await API.post('/api/user/manage', dataToSend);
      const { success, message, data } = res.data;

      if (success) {
        showSuccess('操作成功完成！');
        let newUsers = [...users];
        let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;

        switch (action) {
          case 'delete':
            newUsers[realIdx].deleted = true;
            break;
          case 'disable':
          case 'enable':
            newUsers[realIdx].status = data.status; // 假设API返回了新的状态
            break;
          case 'changeGroup':
            newUsers[realIdx].group = data.group; // 假设API返回了新的用户组信息
            break;
          default:
            console.error('未知的操作类型');
        }

        setUsers(newUsers);
      } else {
        showError(message);
      }
    })();
  };



  const groupOptions = [
    { key: 'default', value: 'default', text: '默认', color: 'var(--czl-grayA)' },
    { key: 'vip', value: 'vip', text: 'VIP', color: 'var(--czl-success-color)' },
    { key: 'svip', value: 'svip', text: '超级VIP', color: 'var(--czl-error-color)' },
  ];

  const groupColor = (userGroup) => {
    const group = groupOptions.find((option) => option.value === userGroup);
    return group ? group.color : 'inherit'; // 如果未找到分组，则返回默认颜色
  };



  const renderStatus = (status) => {
    switch (status) {
      case 1:
        return <Label basic>已激活</Label>;
      case 2:
        return (
          <Label basic color='red'>
            已封禁
          </Label>
        );
      default:
        return (
          <Label basic color='grey'>
            未知状态
          </Label>
        );
    }
  };

  const searchUsers = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load files instead.
      await loadUsers(0);
      setActivePage(1);
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

  return (
    <>
      <Form onSubmit={searchUsers}>
        <Form.Input
          icon='search'
          fluid
          iconPosition='left'
          placeholder='搜索用户的 ID，用户名，显示名称，以及邮箱地址 ...'
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
                sortUser('id');
              }}
            >
              ID
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('username');
              }}
            >
              用户名
            </Table.HeaderCell>
            {/* <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('group');
              }}
            >
              分组
            </Table.HeaderCell> */}
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('quota');
              }}
            >
              统计信息
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('role');
              }}
            >
              用户角色
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('status');
              }}
            >
              状态
            </Table.HeaderCell>
            <Table.HeaderCell>操作</Table.HeaderCell>
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
                      header={user.display_name ? user.display_name : user.username}
                      trigger={<span>{renderText(user.username, 15)}</span>}
                      hoverable
                    />
                  </Table.Cell>
                  {/* <Table.Cell>{renderGroup(user.group)}</Table.Cell> */}
                  <Table.Cell>
                    <Popup content='剩余额度' trigger={<Label basic>{renderQuota(user.quota)}</Label>} />
                    <Popup content='已用额度' trigger={<Label basic>{renderQuota(user.used_quota)}</Label>} />
                    <Popup content='请求次数' trigger={<Label basic>{renderNumber(user.request_count)}</Label>} />
                  </Table.Cell>
                  <Table.Cell>{renderRole(user.role)}</Table.Cell>
                  <Table.Cell>{renderStatus(user.status)}</Table.Cell>
                  <Table.Cell>
                    <div>
                      <Button.Group size={'small'} style={{ marginRight: '10px' }}>
                        <Button
                          positive
                          size={'small'}
                          className={`group-button ${user.group}`}
                          style={{
                            backgroundColor: groupColor(user.group),
                            width: '100px',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                          }}
                        >
                          {user.group}
                        </Button>
                        <Dropdown
                          className="button icon"
                          style={{
                            backgroundColor: groupColor(user.group),
                          }}
                          floating
                          options={groupOptions}
                          trigger={<></>}
                          onChange={(e, { value }) => manageUser(user.username, 'changeGroup', idx, value)}
                        />

                      </Button.Group>
                      <Popup
                        trigger={
                          <Button size='small' negative icon='delete' disabled={user.role === 100} />
                        }
                        on='click'
                        flowing
                        hoverable
                      >
                        <Button
                          negative
                          icon='delete'
                          onClick={() => {
                            manageUser(user.username, 'delete', idx);
                          }}
                        >
                          确认删除{user.username}
                          </Button>
                      </Popup>
                      <Button
                        size={'small'}
                        negative
                        icon={user.status === 1 ? 'ban' : 'check'}
                        style={{
                          backgroundColor: user.status === 1 ? 'var(--czl-warning-color)' : 'var(--czl-success-color)',
                        }}
                        onClick={() => {
                          manageUser(
                            user.username,
                            user.status === 1 ? 'disable' : 'enable',
                            idx
                          );
                        }}
                        disabled={user.role === 100}
                      />
                      <Button
                        size={'small'}
                        icon="edit"
                        style={{ backgroundColor: 'var(--czl-primary-color)' }}
                        as={Link}
                        to={'/user/edit/' + user.id}
                      />

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
                添加新的用户
              </Button>
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
                firstItem={null}  // 不显示第一页按钮
                lastItem={null}  // 不显示最后一页按钮
              />
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
    </>
  );
};

export default UsersTable;
