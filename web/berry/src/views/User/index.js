import { useState, useEffect } from 'react';
import { showError, showSuccess } from 'utils/common';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableContainer from '@mui/material/TableContainer';
import PerfectScrollbar from 'react-perfect-scrollbar';
import TablePagination from '@mui/material/TablePagination';
import LinearProgress from '@mui/material/LinearProgress';
import ButtonGroup from '@mui/material/ButtonGroup';
import Toolbar from '@mui/material/Toolbar';

import { Button, Card, Box, Stack, Container, Typography } from '@mui/material';
import UsersTableRow from './component/TableRow';
import UsersTableHead from './component/TableHead';
import TableToolBar from 'ui-component/TableToolBar';
import { API } from 'utils/api';
import { ITEMS_PER_PAGE } from 'constants';
import { IconRefresh, IconPlus } from '@tabler/icons-react';
import EditeModal from './component/EditModal';

// ----------------------------------------------------------------------
export default function Users() {
  const [users, setUsers] = useState([]);
  const [activePage, setActivePage] = useState(0);
  const [searching, setSearching] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [openModal, setOpenModal] = useState(false);
  const [editUserId, setEditUserId] = useState(0);

  const loadUsers = async (startIdx) => {
    setSearching(true);
    const res = await API.get(`/api/user/?p=${startIdx}`);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setUsers(data);
      } else {
        let newUsers = [...users];
        newUsers.splice(startIdx * ITEMS_PER_PAGE, data.length, ...data);
        setUsers(newUsers);
      }
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const onPaginationChange = (event, activePage) => {
    (async () => {
      if (activePage === Math.ceil(users.length / ITEMS_PER_PAGE)) {
        // In this case we have to load more data and then append them.
        await loadUsers(activePage);
      }
      setActivePage(activePage);
    })();
  };

  const searchUsers = async (event) => {
    event.preventDefault();
    if (searchKeyword === '') {
      await loadUsers(0);
      setActivePage(0);
      return;
    }
    setSearching(true);
    const res = await API.get(`/api/user/search?keyword=${searchKeyword}`);
    const { success, message, data } = res.data;
    if (success) {
      setUsers(data);
      setActivePage(0);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleSearchKeyword = (event) => {
    setSearchKeyword(event.target.value);
  };

  const manageUser = async (username, action, value) => {
    const url = '/api/user/manage';
    let data = { username: username, action: '' };
    let res;
    switch (action) {
      case 'delete':
        data.action = 'delete';
        break;
      case 'status':
        data.action = value === 1 ? 'enable' : 'disable';
        break;
      case 'role':
        data.action = value === true ? 'promote' : 'demote';
        break;
    }

    res = await API.post(url, data);
    const { success, message } = res.data;
    if (success) {
      showSuccess('操作成功完成！');
      await loadUsers(activePage);
    } else {
      showError(message);
    }

    return res.data;
  };

  // 处理刷新
  const handleRefresh = async () => {
    await loadUsers(activePage);
  };

  const handleOpenModal = (userId) => {
    setEditUserId(userId);
    setOpenModal(true);
  };

  const handleCloseModal = () => {
    setOpenModal(false);
    setEditUserId(0);
  };

  const handleOkModal = (status) => {
    if (status === true) {
      handleCloseModal();
      handleRefresh();
    }
  };

  useEffect(() => {
    loadUsers(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []);

  return (
    <>
      <Stack direction="row" alignItems="center" justifyContent="space-between" mb={2.5}>
        <Typography variant="h4">用户</Typography>

        <Button variant="contained" color="primary" startIcon={<IconPlus />} onClick={() => handleOpenModal(0)}>
          新建用户
        </Button>
      </Stack>
      <Card>
        <Box component="form" onSubmit={searchUsers} noValidate sx={{marginTop: 2}}>
          <TableToolBar
            filterName={searchKeyword}
            handleFilterName={handleSearchKeyword}
            placeholder={'搜索用户的ID，用户名，显示名称，以及邮箱地址...'}
          />
        </Box>
        <Toolbar
          sx={{
            textAlign: 'right',
            height: 50,
            display: 'flex',
            justifyContent: 'space-between',
            p: (theme) => theme.spacing(0, 1, 0, 3)
          }}
        >
          <Container>
            <ButtonGroup variant="outlined" aria-label="outlined small primary button group" sx={{marginBottom: 2}}>
              <Button onClick={handleRefresh} startIcon={<IconRefresh width={'18px'} />}>
                刷新
              </Button>
            </ButtonGroup>
          </Container>
        </Toolbar>
        {searching && <LinearProgress />}
        <PerfectScrollbar component="div">
          <TableContainer sx={{ overflow: 'unset' }}>
            <Table sx={{ minWidth: 800 }}>
              <UsersTableHead />
              <TableBody>
                {users.slice(activePage * ITEMS_PER_PAGE, (activePage + 1) * ITEMS_PER_PAGE).map((row) => (
                  <UsersTableRow
                    item={row}
                    manageUser={manageUser}
                    key={row.id}
                    handleOpenModal={handleOpenModal}
                    setModalUserId={setEditUserId}
                  />
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </PerfectScrollbar>
        <TablePagination
          page={activePage}
          component="div"
          count={users.length + (users.length % ITEMS_PER_PAGE === 0 ? 1 : 0)}
          rowsPerPage={ITEMS_PER_PAGE}
          onPageChange={onPaginationChange}
          rowsPerPageOptions={[ITEMS_PER_PAGE]}
        />
      </Card>
      <EditeModal open={openModal} onCancel={handleCloseModal} onOk={handleOkModal} userId={editUserId} />
    </>
  );
}
