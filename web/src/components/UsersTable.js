import React, { useEffect, useState } from 'react';
import { Button, Form, Label, Pagination, Popup, Table } from 'semantic-ui-react';
import { Link } from 'react-router-dom';
import { API, showError, showSuccess } from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';
import { renderText } from '../helpers/render';

function renderRole(role) {
  switch (role) {
    case 1:
      return <Label>Regular user</Label>;
    case 10:
      return <Label color='yellow'>Administrator</Label>;
    case 100:
      return <Label color='orange'>Super Administrator</Label>;
    default:
      return <Label color='red'>Unknown Identity</Label>;
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

  const manageUser = (username, action, idx) => {
    (async () => {
      const res = await API.post('/api/user/manage', {
        username,
        action
      });
      const { success, message } = res.data;
      if (success) {
        showSuccess('Operation completed successfully！');
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
        return <Label basic>Activated</Label>;
      case 2:
        return (
          <Label basic color='red'>
            Banned
          </Label>
        );
      default:
        return (
          <Label basic color='grey'>
            Unknown state
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
      return ('' + a[key]).localeCompare(b[key]);
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
          placeholder='Search for user ID, username, display name, and email address ...'
          value={searchKeyword}
          loading={searching}
          onChange={handleKeywordChange}
        />
      </Form>

      <Table basic>
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
              Username
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('email');
              }}
            >
              email address
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('quota');
              }}
            >
              remaining amount
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('role');
              }}
            >
              user role
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortUser('status');
              }}
            >
              state
            </Table.HeaderCell>
            <Table.HeaderCell>operate</Table.HeaderCell>
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
                      content={user.email ? user.email : 'Email address not found'}
                      key={user.display_name}
                      header={user.display_name ? user.display_name : user.username}
                      trigger={<span>{renderText(user.username, 10)}</span>}
                      hoverable
                    />
                  </Table.Cell>
                  <Table.Cell>{user.email ? renderText(user.email, 30) : 'none'}</Table.Cell>
                  <Table.Cell>{user.quota}</Table.Cell>
                  <Table.Cell>{renderRole(user.role)}</Table.Cell>
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
                        promote
                      </Button>
                      <Button
                        size={'small'}
                        color={'yellow'}
                        onClick={() => {
                          manageUser(user.username, 'demote', idx);
                        }}
                        disabled={user.role === 100}
                      >
                        downgrade
                      </Button>
                      <Popup
                        trigger={
                          <Button size='small' negative disabled={user.role === 100}>
                            delete
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
                          delete users {user.username}
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
                        {user.status === 1 ? 'disabled' : 'enable'}
                      </Button>
                      <Button
                        size={'small'}
                        as={Link}
                        to={'/user/edit/' + user.id}
                        disabled={user.role === 100}
                      >
                        edit
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
                add new user
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
              />
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
    </>
  );
};

export default UsersTable;
