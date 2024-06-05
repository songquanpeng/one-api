import PropTypes from 'prop-types';
import { OutlinedInput, Stack, FormControl, InputLabel, Select, MenuItem } from '@mui/material';
import { LocalizationProvider, DateTimePicker } from '@mui/x-date-pickers';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import dayjs from 'dayjs';
import { StatusType } from './OrderTableRow';
require('dayjs/locale/zh-cn');
// ----------------------------------------------------------------------

export default function OrderTableToolBar({ filterName, handleFilterName }) {
  return (
    <>
      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={{ xs: 3, sm: 2, md: 4 }} padding={'24px'} paddingBottom={'0px'}>
        <FormControl>
          <InputLabel htmlFor="channel-user_id-label">用户ID</InputLabel>
          <OutlinedInput
            id="user_id"
            name="user_id"
            sx={{
              minWidth: '100%'
            }}
            label="用户ID"
            value={filterName.user_id}
            onChange={handleFilterName}
            placeholder="用户ID"
          />
        </FormControl>
        <FormControl>
          <InputLabel htmlFor="channel-trade_no-label">订单号</InputLabel>
          <OutlinedInput
            id="trade_no"
            name="trade_no"
            sx={{
              minWidth: '100%'
            }}
            label="订单号"
            value={filterName.trade_no}
            onChange={handleFilterName}
            placeholder="订单号"
          />
        </FormControl>
        <FormControl>
          <InputLabel htmlFor="channel-gateway_no-label">网关订单号</InputLabel>
          <OutlinedInput
            id="gateway_no"
            name="gateway_no"
            sx={{
              minWidth: '100%'
            }}
            label="网关订单号"
            value={filterName.gateway_no}
            onChange={handleFilterName}
            placeholder="网关订单号"
          />
        </FormControl>

        <FormControl>
          <LocalizationProvider dateAdapter={AdapterDayjs} adapterLocale={'zh-cn'}>
            <DateTimePicker
              label="起始时间"
              ampm={false}
              name="start_timestamp"
              value={filterName.start_timestamp === 0 ? null : dayjs.unix(filterName.start_timestamp)}
              onChange={(value) => {
                if (value === null) {
                  handleFilterName({ target: { name: 'start_timestamp', value: 0 } });
                  return;
                }
                handleFilterName({ target: { name: 'start_timestamp', value: value.unix() } });
              }}
              slotProps={{
                actionBar: {
                  actions: ['clear', 'today', 'accept']
                }
              }}
            />
          </LocalizationProvider>
        </FormControl>

        <FormControl>
          <LocalizationProvider dateAdapter={AdapterDayjs} adapterLocale={'zh-cn'}>
            <DateTimePicker
              label="结束时间"
              name="end_timestamp"
              ampm={false}
              value={filterName.end_timestamp === 0 ? null : dayjs.unix(filterName.end_timestamp)}
              onChange={(value) => {
                if (value === null) {
                  handleFilterName({ target: { name: 'end_timestamp', value: 0 } });
                  return;
                }
                handleFilterName({ target: { name: 'end_timestamp', value: value.unix() } });
              }}
              slotProps={{
                actionBar: {
                  actions: ['clear', 'today', 'accept']
                }
              }}
            />
          </LocalizationProvider>
        </FormControl>
        <FormControl sx={{ minWidth: '22%' }}>
          <InputLabel htmlFor="channel-status-label">状态</InputLabel>
          <Select
            id="channel-type-label"
            label="状态"
            value={filterName.status}
            name="status"
            onChange={handleFilterName}
            sx={{
              minWidth: '100%'
            }}
            MenuProps={{
              PaperProps: {
                style: {
                  maxHeight: 200
                }
              }
            }}
          >
            {Object.values(StatusType).map((option) => {
              return (
                <MenuItem key={option.value} value={option.value}>
                  {option.name}
                </MenuItem>
              );
            })}
          </Select>
        </FormControl>
      </Stack>
    </>
  );
}

OrderTableToolBar.propTypes = {
  filterName: PropTypes.object,
  handleFilterName: PropTypes.func
};
