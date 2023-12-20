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
import RedemptionTableRow from './component/TableRow';
import RedemptionTableHead from './component/TableHead';
import TableToolBar from 'ui-component/TableToolBar';
import { API } from 'utils/api';
import { ITEMS_PER_PAGE } from 'constants';
import { IconRefresh, IconPlus } from '@tabler/icons-react';
import EditeModal from './component/EditModal';

// ----------------------------------------------------------------------
export default function Redemption() {
  const [redemptions, setRedemptions] = useState([]);
  const [activePage, setActivePage] = useState(0);
  const [searching, setSearching] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [openModal, setOpenModal] = useState(false);
  const [editRedemptionId, setEditRedemptionId] = useState(0);

  const loadRedemptions = async (startIdx) => {
    setSearching(true);
    const res = await API.get(`/api/redemption/?p=${startIdx}`);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setRedemptions(data);
      } else {
        let newRedemptions = [...redemptions];
        newRedemptions.splice(startIdx * ITEMS_PER_PAGE, data.length, ...data);
        setRedemptions(newRedemptions);
      }
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const onPaginationChange = (event, activePage) => {
    (async () => {
      if (activePage === Math.ceil(redemptions.length / ITEMS_PER_PAGE)) {
        // In this case we have to load more data and then append them.
        await loadRedemptions(activePage);
      }
      setActivePage(activePage);
    })();
  };

  const searchRedemptions = async (event) => {
    event.preventDefault();
    if (searchKeyword === '') {
      await loadRedemptions(0);
      setActivePage(0);
      return;
    }
    setSearching(true);
    const res = await API.get(`/api/redemption/search?keyword=${searchKeyword}`);
    const { success, message, data } = res.data;
    if (success) {
      setRedemptions(data);
      setActivePage(0);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleSearchKeyword = (event) => {
    setSearchKeyword(event.target.value);
  };

  const manageRedemptions = async (id, action, value) => {
    const url = '/api/redemption/';
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(url + id);
        break;
      case 'status':
        res = await API.put(url + '?status_only=true', {
          ...data,
          status: value
        });
        break;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess('操作成功完成！');
      if (action === 'delete') {
        await loadRedemptions(0);
      }
    } else {
      showError(message);
    }

    return res.data;
  };

  // 处理刷新
  const handleRefresh = async () => {
    await loadRedemptions(0);
    setActivePage(0);
    setSearchKeyword('');
  };

  const handleOpenModal = (redemptionId) => {
    setEditRedemptionId(redemptionId);
    setOpenModal(true);
  };

  const handleCloseModal = () => {
    setOpenModal(false);
    setEditRedemptionId(0);
  };

  const handleOkModal = (status) => {
    if (status === true) {
      handleCloseModal();
      handleRefresh();
    }
  };

  useEffect(() => {
    loadRedemptions(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []);

  return (
    <>
      <Stack direction="row" alignItems="center" justifyContent="space-between" mb={5}>
        <Typography variant="h4">兑换</Typography>

        <Button variant="contained" color="primary" startIcon={<IconPlus />} onClick={() => handleOpenModal(0)}>
          新建兑换码
        </Button>
      </Stack>
      <Card>
        <Box component="form" onSubmit={searchRedemptions} noValidate>
          <TableToolBar filterName={searchKeyword} handleFilterName={handleSearchKeyword} placeholder={'搜索兑换码的ID和名称...'} />
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
              <RedemptionTableHead />
              <TableBody>
                {redemptions.slice(activePage * ITEMS_PER_PAGE, (activePage + 1) * ITEMS_PER_PAGE).map((row) => (
                  <RedemptionTableRow
                    item={row}
                    manageRedemption={manageRedemptions}
                    key={row.id}
                    handleOpenModal={handleOpenModal}
                    setModalRedemptionId={setEditRedemptionId}
                  />
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </PerfectScrollbar>
        <TablePagination
          page={activePage}
          component="div"
          count={redemptions.length + (redemptions.length % ITEMS_PER_PAGE === 0 ? 1 : 0)}
          rowsPerPage={ITEMS_PER_PAGE}
          onPageChange={onPaginationChange}
          rowsPerPageOptions={[ITEMS_PER_PAGE]}
        />
      </Card>
      <EditeModal open={openModal} onCancel={handleCloseModal} onOk={handleOkModal} redemptiondId={editRedemptionId} />
    </>
  );
}
