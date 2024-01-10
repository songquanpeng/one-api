import { useState, useEffect } from 'react';
import { showError, showSuccess } from 'utils/common';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableContainer from '@mui/material/TableContainer';
import PerfectScrollbar from 'react-perfect-scrollbar';
import TablePagination from '@mui/material/TablePagination';
import LinearProgress from '@mui/material/LinearProgress';
import Alert from '@mui/material/Alert';
import ButtonGroup from '@mui/material/ButtonGroup';
import Toolbar from '@mui/material/Toolbar';

import { Button, Card, Box, Stack, Container, Typography } from '@mui/material';
import TokensTableRow from './component/TableRow';
import TokenTableHead from './component/TableHead';
import TableToolBar from 'ui-component/TableToolBar';
import { API } from 'utils/api';
import { ITEMS_PER_PAGE } from 'constants';
import { IconRefresh, IconPlus } from '@tabler/icons-react';
import EditeModal from './component/EditModal';
import { useSelector } from 'react-redux';

export default function Token() {
  const [tokens, setTokens] = useState([]);
  const [activePage, setActivePage] = useState(0);
  const [searching, setSearching] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [openModal, setOpenModal] = useState(false);
  const [editTokenId, setEditTokenId] = useState(0);
  const siteInfo = useSelector((state) => state.siteInfo);

  const loadTokens = async (startIdx) => {
    setSearching(true);
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
    setSearching(false);
  };

  useEffect(() => {
    loadTokens(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []);

  const onPaginationChange = (event, activePage) => {
    (async () => {
      if (activePage === Math.ceil(tokens.length / ITEMS_PER_PAGE)) {
        // In this case we have to load more data and then append them.
        await loadTokens(activePage);
      }
      setActivePage(activePage);
    })();
  };

  const searchTokens = async (event) => {
    event.preventDefault();
    if (searchKeyword === '') {
      await loadTokens(0);
      setActivePage(0);
      return;
    }
    setSearching(true);
    const res = await API.get(`/api/token/search?keyword=${searchKeyword}`);
    const { success, message, data } = res.data;
    if (success) {
      setTokens(data);
      setActivePage(0);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleSearchKeyword = (event) => {
    setSearchKeyword(event.target.value);
  };

  const manageToken = async (id, action, value) => {
    const url = '/api/token/';
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(url + id);
        break;
      case 'status':
        res = await API.put(url + `?status_only=true`, {
          ...data,
          status: value
        });
        break;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess('操作成功完成！');
      if (action === 'delete') {
        await handleRefresh();
      }
    } else {
      showError(message);
    }

    return res.data;
  };

  // 处理刷新
  const handleRefresh = async () => {
    await loadTokens(activePage);
  };

  const handleOpenModal = (tokenId) => {
    setEditTokenId(tokenId);
    setOpenModal(true);
  };

  const handleCloseModal = () => {
    setOpenModal(false);
    setEditTokenId(0);
  };

  const handleOkModal = (status) => {
    if (status === true) {
      handleCloseModal();
      handleRefresh();
    }
  };

  return (
    <>
      <Stack direction="row" alignItems="center" justifyContent="space-between" mb={5}>
        <Typography variant="h4">令牌</Typography>

        <Button
          variant="contained"
          color="primary"
          onClick={() => {
            handleOpenModal(0);
          }}
          startIcon={<IconPlus />}
        >
          新建令牌
        </Button>
      </Stack>
      <Stack mb={5}>
        <Alert severity="info">
          将 OpenAI API 基础地址 https://api.openai.com 替换为 <b>{siteInfo.server_address}</b>，复制下面的密钥即可使用
        </Alert>
      </Stack>
      <Card>
        <Box component="form" onSubmit={searchTokens} noValidate>
          <TableToolBar filterName={searchKeyword} handleFilterName={handleSearchKeyword} placeholder={'搜索令牌的名称...'} />
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
            <ButtonGroup variant="outlined" aria-label="outlined small primary button group">
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
              <TokenTableHead />
              <TableBody>
                {tokens.slice(activePage * ITEMS_PER_PAGE, (activePage + 1) * ITEMS_PER_PAGE).map((row) => (
                  <TokensTableRow
                    item={row}
                    manageToken={manageToken}
                    key={row.id}
                    handleOpenModal={handleOpenModal}
                    setModalTokenId={setEditTokenId}
                  />
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </PerfectScrollbar>
        <TablePagination
          page={activePage}
          component="div"
          count={tokens.length + (tokens.length % ITEMS_PER_PAGE === 0 ? 1 : 0)}
          rowsPerPage={ITEMS_PER_PAGE}
          onPageChange={onPaginationChange}
          rowsPerPageOptions={[ITEMS_PER_PAGE]}
        />
      </Card>
      <EditeModal open={openModal} onCancel={handleCloseModal} onOk={handleOkModal} tokenId={editTokenId} />
    </>
  );
}
