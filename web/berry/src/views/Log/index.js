import { useState, useEffect } from 'react';
import { showError } from 'utils/common';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableContainer from '@mui/material/TableContainer';
import PerfectScrollbar from 'react-perfect-scrollbar';
import TablePagination from '@mui/material/TablePagination';
import LinearProgress from '@mui/material/LinearProgress';
import ButtonGroup from '@mui/material/ButtonGroup';
import Toolbar from '@mui/material/Toolbar';

import { Button, Card, Stack, Container, Typography, Box } from '@mui/material';
import LogTableRow from './component/TableRow';
import LogTableHead from './component/TableHead';
import TableToolBar from './component/TableToolBar';
import { API } from 'utils/api';
import { isAdmin } from 'utils/common';
import { ITEMS_PER_PAGE } from 'constants';
import { IconRefresh, IconSearch } from '@tabler/icons-react';

export default function Log() {
  const originalKeyword = {
    p: 0,
    username: '',
    token_name: '',
    model_name: '',
    start_timestamp: 0,
    end_timestamp: new Date().getTime() / 1000 + 3600,
    type: 0,
    channel: ''
  };
  const [logs, setLogs] = useState([]);
  const [activePage, setActivePage] = useState(0);
  const [searching, setSearching] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState(originalKeyword);
  const [initPage, setInitPage] = useState(true);
  const userIsAdmin = isAdmin();

  const loadLogs = async (startIdx) => {
    setSearching(true);
    const url = userIsAdmin ? '/api/log/' : '/api/log/self/';
    const query = searchKeyword;

    query.p = startIdx;
    if (!userIsAdmin) {
      delete query.username;
      delete query.channel;
    }
    const res = await API.get(url, { params: query });
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setLogs(data);
      } else {
        let newLogs = [...logs];
        newLogs.splice(startIdx * ITEMS_PER_PAGE, data.length, ...data);
        setLogs(newLogs);
      }
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const onPaginationChange = (event, activePage) => {
    (async () => {
      if (activePage === Math.ceil(logs.length / ITEMS_PER_PAGE)) {
        // In this case we have to load more data and then append them.
        await loadLogs(activePage);
      }
      setActivePage(activePage);
    })();
  };

  const searchLogs = async (event) => {
    event.preventDefault();
    await loadLogs(0);
    setActivePage(0);
    return;
  };

  const handleSearchKeyword = (event) => {
    setSearchKeyword({ ...searchKeyword, [event.target.name]: event.target.value });
  };

  // 处理刷新
  const handleRefresh = () => {
    setInitPage(true);
  };

  useEffect(() => {
    setSearchKeyword(originalKeyword);
    setActivePage(0);
    loadLogs(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
    setInitPage(false);
  }, [initPage]);

  return (
    <>
      <Stack direction="row" alignItems="center" justifyContent="space-between" mb={5}>
        <Typography variant="h4">日志</Typography>
      </Stack>
      <Card>
        <Box component="form" onSubmit={searchLogs} noValidate>
          <TableToolBar filterName={searchKeyword} handleFilterName={handleSearchKeyword} userIsAdmin={userIsAdmin} />
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

              <Button onClick={searchLogs} startIcon={<IconSearch width={'18px'} />}>
                搜索
              </Button>
            </ButtonGroup>
          </Container>
        </Toolbar>
        {searching && <LinearProgress />}
        <PerfectScrollbar component="div">
          <TableContainer sx={{ overflow: 'unset' }}>
            <Table sx={{ minWidth: 800 }}>
              <LogTableHead userIsAdmin={userIsAdmin} />
              <TableBody>
                {logs.slice(activePage * ITEMS_PER_PAGE, (activePage + 1) * ITEMS_PER_PAGE).map((row, index) => (
                  <LogTableRow item={row} key={`${row.id}_${index}`} userIsAdmin={userIsAdmin} />
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </PerfectScrollbar>
        <TablePagination
          page={activePage}
          component="div"
          count={logs.length + (logs.length % ITEMS_PER_PAGE === 0 ? 1 : 0)}
          rowsPerPage={ITEMS_PER_PAGE}
          onPageChange={onPaginationChange}
          rowsPerPageOptions={[ITEMS_PER_PAGE]}
        />
      </Card>
    </>
  );
}
