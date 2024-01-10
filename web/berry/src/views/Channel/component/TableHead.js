import { TableCell, TableHead, TableRow } from '@mui/material';

const ChannelTableHead = () => {
  return (
    <TableHead>
      <TableRow>
        <TableCell>ID</TableCell>
        <TableCell>名称</TableCell>
        <TableCell>分组</TableCell>
        <TableCell>类型</TableCell>
        <TableCell>状态</TableCell>
        <TableCell>响应时间</TableCell>
        <TableCell>余额</TableCell>
        <TableCell>优先级</TableCell>
        <TableCell>操作</TableCell>
      </TableRow>
    </TableHead>
  );
};

export default ChannelTableHead;
