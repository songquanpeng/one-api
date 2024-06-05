import PropTypes from 'prop-types';
import { OutlinedInput, Stack, FormControl, InputLabel, Select, MenuItem } from '@mui/material';
import { PaymentType } from '../type/Config';
require('dayjs/locale/zh-cn');
// ----------------------------------------------------------------------

export default function TableToolBar({ filterName, handleFilterName }) {
  return (
    <>
      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={{ xs: 3, sm: 2, md: 4 }} padding={'24px'} paddingBottom={'0px'}>
        <FormControl>
          <InputLabel htmlFor="channel-name-label">名称</InputLabel>
          <OutlinedInput
            id="name"
            name="name"
            sx={{
              minWidth: '100%'
            }}
            label="名称"
            value={filterName.name}
            onChange={handleFilterName}
            placeholder="名称"
          />
        </FormControl>
        <FormControl>
          <InputLabel htmlFor="channel-uuid-label">UUID</InputLabel>
          <OutlinedInput
            id="uuid"
            name="uuid"
            sx={{
              minWidth: '100%'
            }}
            label="模型名称"
            value={filterName.uuid}
            onChange={handleFilterName}
            placeholder="UUID"
          />
        </FormControl>
        <FormControl sx={{ minWidth: '22%' }}>
          <InputLabel htmlFor="channel-type-label">类型</InputLabel>
          <Select
            id="channel-type-label"
            label="类型"
            value={filterName.type}
            name="type"
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
            {Object.entries(PaymentType).map(([value, text]) => (
              <MenuItem key={value} value={value}>
                {text}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      </Stack>
    </>
  );
}

TableToolBar.propTypes = {
  filterName: PropTypes.object,
  handleFilterName: PropTypes.func
};
