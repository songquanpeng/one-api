import { TableCell, TableHead, TableRow } from '@mui/material';

const RedemptionTableHead = () => {
  return (
    <TableHead>
      <TableRow>
        <TableCell>ID</TableCell>
        <TableCell>名称</TableCell>
        <TableCell>状态</TableCell>
        <TableCell>额度</TableCell>
        <TableCell>创建时间</TableCell>
        <TableCell>兑换时间</TableCell>
        <TableCell>操作</TableCell>
      </TableRow>
    </TableHead>
  );
};

export default RedemptionTableHead;
