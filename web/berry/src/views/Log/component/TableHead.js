import PropTypes from 'prop-types';
import { TableCell, TableHead, TableRow } from '@mui/material';

const LogTableHead = ({ userIsAdmin }) => {
  return (
    <TableHead>
      <TableRow>
        <TableCell>时间</TableCell>
        {userIsAdmin && <TableCell>渠道</TableCell>}
        {userIsAdmin && <TableCell>用户</TableCell>}
        <TableCell>令牌</TableCell>
        <TableCell>类型</TableCell>
        <TableCell>模型</TableCell>
        <TableCell>提示</TableCell>
        <TableCell>补全</TableCell>
        <TableCell>额度</TableCell>
        <TableCell>详情</TableCell>
      </TableRow>
    </TableHead>
  );
};

export default LogTableHead;

LogTableHead.propTypes = {
  userIsAdmin: PropTypes.bool
};
