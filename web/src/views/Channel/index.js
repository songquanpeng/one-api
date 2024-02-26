import { useState, useEffect } from 'react';
import { showError, showSuccess, showInfo } from 'utils/common';

import { useTheme } from '@mui/material/styles';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableContainer from '@mui/material/TableContainer';
import PerfectScrollbar from 'react-perfect-scrollbar';
import TablePagination from '@mui/material/TablePagination';
import LinearProgress from '@mui/material/LinearProgress';
import ButtonGroup from '@mui/material/ButtonGroup';
import Toolbar from '@mui/material/Toolbar';
import useMediaQuery from '@mui/material/useMediaQuery';

import { Button, IconButton, Card, Box, Stack, Container, Typography, Divider } from '@mui/material';
import ChannelTableRow from './component/TableRow';
import KeywordTableHead from 'ui-component/TableHead';
import { API } from 'utils/api';
import { IconRefresh, IconHttpDelete, IconPlus, IconMenu2, IconBrandSpeedtest, IconCoinYuan, IconSearch } from '@tabler/icons-react';
import EditeModal from './component/EditModal';
import { ITEMS_PER_PAGE } from 'constants';
import TableToolBar from './component/TableToolBar';
import BatchModal from './component/BatchModal';

const originalKeyword = {
  type: 0,
  status: 0,
  name: '',
  group: '',
  models: '',
  key: '',
  test_model: '',
  other: ''
};

export async function fetchChannelData(page, rowsPerPage, keyword, order, orderBy) {
  try {
    if (orderBy) {
      orderBy = order === 'desc' ? '-' + orderBy : orderBy;
    }
    const res = await API.get(`/api/channel/`, {
      params: {
        page: page + 1,
        size: rowsPerPage,
        order: orderBy,
        ...keyword
      }
    });
    const { success, message, data } = res.data;
    if (success) {
      return data;
    } else {
      showError(message);
    }
  } catch (error) {
    console.error(error);
  }

  return false;
}

// ----------------------------------------------------------------------
// CHANNEL_OPTIONS,
export default function ChannelPage() {
  const [page, setPage] = useState(0);
  const [order, setOrder] = useState('desc');
  const [orderBy, setOrderBy] = useState('id');
  const [rowsPerPage, setRowsPerPage] = useState(ITEMS_PER_PAGE);
  const [listCount, setListCount] = useState(0);
  const [searching, setSearching] = useState(false);
  const [channels, setChannels] = useState([]);
  const [refreshFlag, setRefreshFlag] = useState(false);

  const [groupOptions, setGroupOptions] = useState([]);
  const [toolBarValue, setToolBarValue] = useState(originalKeyword);
  const [searchKeyword, setSearchKeyword] = useState(originalKeyword);

  const theme = useTheme();
  const matchUpMd = useMediaQuery(theme.breakpoints.up('sm'));
  const [openModal, setOpenModal] = useState(false);
  const [editChannelId, setEditChannelId] = useState(0);
  const [openBatchModal, setOpenBatchModal] = useState(false);

  const handleSort = (event, id) => {
    const isAsc = orderBy === id && order === 'asc';
    if (id !== '') {
      setOrder(isAsc ? 'desc' : 'asc');
      setOrderBy(id);
    }
  };

  const handleChangePage = (event, newPage) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event) => {
    setPage(0);
    setRowsPerPage(parseInt(event.target.value, 10));
  };

  const searchChannels = async () => {
    // event.preventDefault();
    // const formData = new FormData(event.target);
    setPage(0);
    setSearchKeyword(toolBarValue);
  };

  const handleToolBarValue = (event) => {
    setToolBarValue({ ...toolBarValue, [event.target.name]: event.target.value });
  };

  const manageChannel = async (id, action, value) => {
    const url = '/api/channel/';
    let data = { id };
    let res;

    try {
      switch (action) {
        case 'delete':
          res = await API.delete(url + id);
          break;
        case 'status':
          res = await API.put(url, {
            ...data,
            status: value
          });
          break;
        case 'priority':
          if (value === '') {
            return;
          }
          res = await API.put(url, {
            ...data,
            priority: parseInt(value)
          });
          break;
        case 'test':
          res = await API.get(url + `test/${id}`);
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
    } catch (error) {
      return;
    }
  };

  // 处理刷新
  const handleRefresh = async () => {
    setOrderBy('id');
    setOrder('desc');
    setToolBarValue(originalKeyword);
    setSearchKeyword(originalKeyword);
    setRefreshFlag(!refreshFlag);
  };

  // 处理测试所有启用渠道
  const testAllChannels = async () => {
    try {
      const res = await API.get(`/api/channel/test`);
      const { success, message } = res.data;
      if (success) {
        showInfo('已成功开始测试所有通道，请刷新页面查看结果。');
      } else {
        showError(message);
      }
    } catch (error) {
      return;
    }
  };

  // 处理删除所有禁用渠道
  const deleteAllDisabledChannels = async () => {
    try {
      const res = await API.delete(`/api/channel/disabled`);
      const { success, message, data } = res.data;
      if (success) {
        showSuccess(`已删除所有禁用渠道，共计 ${data} 个`);
        await handleRefresh();
      } else {
        showError(message);
      }
    } catch (error) {
      return;
    }
  };

  // 处理更新所有启用渠道余额
  const updateAllChannelsBalance = async () => {
    setSearching(true);
    try {
      const res = await API.get(`/api/channel/update_balance`);
      const { success, message } = res.data;
      if (success) {
        showInfo('已更新完毕所有已启用通道余额！');
      } else {
        showError(message);
      }
    } catch (error) {
      console.log(error);
    }

    setSearching(false);
  };

  const handleOpenModal = (channelId) => {
    setEditChannelId(channelId);
    setOpenModal(true);
  };

  const handleCloseModal = () => {
    setOpenModal(false);
    setEditChannelId(0);
  };

  const handleOkModal = (status) => {
    if (status === true) {
      handleCloseModal();
      handleRefresh();
    }
  };

  const fetchData = async (page, rowsPerPage, keyword, order, orderBy) => {
    setSearching(true);
    const data = await fetchChannelData(page, rowsPerPage, keyword, order, orderBy);

    if (data) {
      setListCount(data.total_count);
      setChannels(data.data);
    }
    setSearching(false);
  };

  const fetchGroups = async () => {
    try {
      let res = await API.get(`/api/group/`);
      setGroupOptions(res.data.data);
    } catch (error) {
      showError(error.message);
    }
  };

  useEffect(() => {
    fetchData(page, rowsPerPage, searchKeyword, order, orderBy);
  }, [page, rowsPerPage, searchKeyword, order, orderBy, refreshFlag]);

  useEffect(() => {
    fetchGroups().then();
  }, []);

  return (
    <>
      <Stack direction="row" alignItems="center" justifyContent="space-between" mb={5}>
        <Typography variant="h4">渠道</Typography>

        <ButtonGroup variant="contained" aria-label="outlined small primary button group">
          <Button color="primary" startIcon={<IconPlus />} onClick={() => handleOpenModal(0)}>
            新建渠道
          </Button>
          <Button color="primary" startIcon={<IconMenu2 />} onClick={() => setOpenBatchModal(true)}>
            批量处理
          </Button>
        </ButtonGroup>
      </Stack>
      <Card>
        <Box component="form" noValidate>
          <TableToolBar filterName={toolBarValue} handleFilterName={handleToolBarValue} groupOptions={groupOptions} />
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
            {matchUpMd ? (
              <ButtonGroup variant="outlined" aria-label="outlined small primary button group">
                <Button onClick={handleRefresh} startIcon={<IconRefresh width={'18px'} />}>
                  刷新/清除搜索条件
                </Button>
                <Button onClick={searchChannels} startIcon={<IconSearch width={'18px'} />}>
                  搜索
                </Button>
                <Button onClick={testAllChannels} startIcon={<IconBrandSpeedtest width={'18px'} />}>
                  测试启用渠道
                </Button>
                <Button onClick={updateAllChannelsBalance} startIcon={<IconCoinYuan width={'18px'} />}>
                  更新启用余额
                </Button>
                <Button onClick={deleteAllDisabledChannels} startIcon={<IconHttpDelete width={'18px'} />}>
                  删除禁用渠道
                </Button>
              </ButtonGroup>
            ) : (
              <Stack
                direction="row"
                spacing={1}
                divider={<Divider orientation="vertical" flexItem />}
                justifyContent="space-around"
                alignItems="center"
              >
                <IconButton onClick={handleRefresh} size="large">
                  <IconRefresh />
                </IconButton>
                <IconButton onClick={searchChannels} size="large">
                  <IconSearch />
                </IconButton>
                <IconButton onClick={testAllChannels} size="large">
                  <IconBrandSpeedtest />
                </IconButton>
                <IconButton onClick={updateAllChannelsBalance} size="large">
                  <IconCoinYuan />
                </IconButton>
                <IconButton onClick={deleteAllDisabledChannels} size="large">
                  <IconHttpDelete />
                </IconButton>
              </Stack>
            )}
          </Container>
        </Toolbar>
        {searching && <LinearProgress />}
        <PerfectScrollbar component="div">
          <TableContainer sx={{ overflow: 'unset' }}>
            <Table sx={{ minWidth: 800 }}>
              <KeywordTableHead
                order={order}
                orderBy={orderBy}
                onRequestSort={handleSort}
                headLabel={[
                  { id: 'collapse', label: '', disableSort: true },
                  { id: 'id', label: 'ID', disableSort: false },
                  { id: 'name', label: '名称', disableSort: false },
                  { id: 'group', label: '分组', disableSort: true },
                  { id: 'type', label: '类型', disableSort: false },
                  { id: 'status', label: '状态', disableSort: false },
                  { id: 'response_time', label: '响应时间', disableSort: false },
                  { id: 'balance', label: '余额', disableSort: false },
                  { id: 'priority', label: '优先级', disableSort: false },
                  { id: 'action', label: '操作', disableSort: true }
                ]}
              />
              <TableBody>
                {channels.map((row) => (
                  <ChannelTableRow
                    item={row}
                    manageChannel={manageChannel}
                    key={row.id}
                    handleOpenModal={handleOpenModal}
                    setModalChannelId={setEditChannelId}
                  />
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </PerfectScrollbar>
        <TablePagination
          page={page}
          component="div"
          count={listCount}
          rowsPerPage={rowsPerPage}
          onPageChange={handleChangePage}
          rowsPerPageOptions={[10, 25, 30]}
          onRowsPerPageChange={handleChangeRowsPerPage}
          showFirstButton
          showLastButton
        />
      </Card>
      <EditeModal open={openModal} onCancel={handleCloseModal} onOk={handleOkModal} channelId={editChannelId} groupOptions={groupOptions} />
      <BatchModal open={openBatchModal} setOpen={setOpenBatchModal} />
    </>
  );
}
