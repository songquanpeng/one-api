import { TableCell, TableHead, TableRow } from '@mui/material';

const TokenTableHead = () => {
  return (
    <TableHead>
      <TableRow>
        <TableCell>名称</TableCell>
        <TableCell>状态</TableCell>
        <TableCell>已用额度</TableCell>
        <TableCell>剩余额度</TableCell>
        <TableCell>创建时间</TableCell>
        <TableCell>过期时间</TableCell>
        <TableCell>操作</TableCell>
      </TableRow>
    </TableHead>
  );
};

export default TokenTableHead;
