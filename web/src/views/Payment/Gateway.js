import { useState, useEffect, useCallback } from 'react';
import { showError, trims, showSuccess } from 'utils/common';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableContainer from '@mui/material/TableContainer';
import PerfectScrollbar from 'react-perfect-scrollbar';
import TablePagination from '@mui/material/TablePagination';
import LinearProgress from '@mui/material/LinearProgress';
import ButtonGroup from '@mui/material/ButtonGroup';
import Toolbar from '@mui/material/Toolbar';

import { Button, Card, Stack, Container, Typography, Box } from '@mui/material';
import PaymentTableRow from './component/TableRow';
import KeywordTableHead from 'ui-component/TableHead';
import TableToolBar from './component/TableToolBar';
import EditeModal from './component/EditModal';
import { API } from 'utils/api';
import { ITEMS_PER_PAGE } from 'constants';
import { IconRefresh, IconSearch, IconPlus } from '@tabler/icons-react';

export default function Gateway() {
  const originalKeyword = {
    p: 0,
    type: '',
    name: '',
    uuid: '',
    currency: ''
  };

  const [page, setPage] = useState(0);
  const [order, setOrder] = useState('desc');
  const [orderBy, setOrderBy] = useState('created_at');
  const [rowsPerPage, setRowsPerPage] = useState(ITEMS_PER_PAGE);
  const [listCount, setListCount] = useState(0);
  const [searching, setSearching] = useState(false);
  const [toolBarValue, setToolBarValue] = useState(originalKeyword);
  const [searchKeyword, setSearchKeyword] = useState(originalKeyword);
  const [refreshFlag, setRefreshFlag] = useState(false);
  const [openModal, setOpenModal] = useState(false);
  const [editPaymentId, setEditPaymentId] = useState(0);

  const [payment, setPayment] = useState([]);

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

  const handleOpenModal = (channelId) => {
    setEditPaymentId(channelId);
    setOpenModal(true);
  };

  const handleCloseModal = () => {
    setOpenModal(false);
    setEditPaymentId(0);
  };

  const handleOkModal = (status) => {
    if (status === true) {
      handleCloseModal();
      handleRefresh();
    }
  };

  const handleChangeRowsPerPage = (event) => {
    setPage(0);
    setRowsPerPage(parseInt(event.target.value, 10));
  };

  const search = async () => {
    setPage(0);
    setSearchKeyword(toolBarValue);
  };

  const handleToolBarValue = (event) => {
    setToolBarValue({ ...toolBarValue, [event.target.name]: event.target.value });
  };

  const managePayment = async (id, action, value) => {
    const url = '/api/payment/';
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
            enable: value
          });
          break;
        case 'sort':
          res = await API.put(url, {
            ...data,
            sort: value
          });
          break;
      }
      const { success, message } = res.data;
      if (success) {
        showSuccess('操作成功完成！');
        await handleRefresh();
      } else {
        showError(message);
      }

      return res.data;
    } catch (error) {
      return;
    }
  };

  const fetchData = useCallback(async (page, rowsPerPage, keyword, order, orderBy) => {
    setSearching(true);
    keyword = trims(keyword);
    try {
      if (orderBy) {
        orderBy = order === 'desc' ? '-' + orderBy : orderBy;
      }
      const res = await API.get('/api/payment/', {
        params: {
          page: page + 1,
          size: rowsPerPage,
          order: orderBy,
          ...keyword
        }
      });
      const { success, message, data } = res.data;
      if (success) {
        setListCount(data.total_count);
        setPayment(data.data);
      } else {
        showError(message);
      }
    } catch (error) {
      console.error(error);
    }
    setSearching(false);
  }, []);

  // 处理刷新
  const handleRefresh = async () => {
    setOrderBy('created_at');
    setOrder('desc');
    setToolBarValue(originalKeyword);
    setSearchKeyword(originalKeyword);
    setRefreshFlag(!refreshFlag);
  };

  useEffect(() => {
    fetchData(page, rowsPerPage, searchKeyword, order, orderBy);
  }, [page, rowsPerPage, searchKeyword, order, orderBy, fetchData, refreshFlag]);

  return (
    <>
      <Stack direction="row" alignItems="center" justifyContent="space-between" mb={5}>
        <Typography variant="h4">支付网关</Typography>
        <Button variant="contained" color="primary" startIcon={<IconPlus />} onClick={() => handleOpenModal(0)}>
          新建支付
        </Button>
      </Stack>
      <Card>
        <Box component="form" noValidate>
          <TableToolBar filterName={toolBarValue} handleFilterName={handleToolBarValue} />
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
                刷新/清除搜索条件
              </Button>

              <Button onClick={search} startIcon={<IconSearch width={'18px'} />}>
                搜索
              </Button>
            </ButtonGroup>
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
                  {
                    id: 'id',
                    label: 'ID',
                    disableSort: false
                  },
                  {
                    id: 'uuid',
                    label: 'UUID',
                    disableSort: false
                  },
                  {
                    id: 'name',
                    label: '名称',
                    disableSort: true
                  },
                  {
                    id: 'type',
                    label: '类型',
                    disableSort: false
                  },
                  {
                    id: 'icon',
                    label: '图标',
                    disableSort: true
                  },
                  {
                    id: 'fixed_fee',
                    label: '固定手续费',
                    disableSort: true
                  },
                  {
                    id: 'percent_fee',
                    label: '百分比手续费',
                    disableSort: true
                  },
                  {
                    id: 'sort',
                    label: '排序',
                    disableSort: false
                  },
                  {
                    id: 'enable',
                    label: '启用',
                    disableSort: false
                  },
                  {
                    id: 'created_at',
                    label: '创建时间',
                    disableSort: false
                  },
                  {
                    id: 'action',
                    label: '操作',
                    disableSort: true
                  }
                ]}
              />
              <TableBody>
                {payment.map((row, index) => (
                  <PaymentTableRow
                    item={row}
                    key={`${row.id}_${index}`}
                    managePayment={managePayment}
                    handleOpenModal={handleOpenModal}
                    setModalPaymentId={setEditPaymentId}
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
        <EditeModal open={openModal} onCancel={handleCloseModal} onOk={handleOkModal} paymentId={editPaymentId} />
      </Card>
    </>
  );
}
