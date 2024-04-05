import PropTypes from 'prop-types';
import { Box, Typography, TableRow, TableCell } from '@mui/material';

const TableNoData = ({ message = '暂无数据' }) => {
  return (
    <TableRow>
      <TableCell colSpan={1000}>
        <Box
          sx={{
            minHeight: '490px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}
        >
          <Typography variant="h3" color={'#697586'}>
            {message}
          </Typography>
        </Box>
      </TableCell>
    </TableRow>
  );
};
export default TableNoData;

TableNoData.propTypes = {
  message: PropTypes.string
};
