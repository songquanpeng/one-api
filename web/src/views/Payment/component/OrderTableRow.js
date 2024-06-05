import PropTypes from 'prop-types';

import { TableRow, TableCell } from '@mui/material';

import { timestamp2string } from 'utils/common';
import Label from 'ui-component/Label';

const StatusType = {
  pending: { name: '待支付', value: 'pending', color: 'primary' },
  success: { name: '支付成功', value: 'success', color: 'success' },
  failed: { name: '支付失败', value: 'failed', color: 'error' },
  closed: { name: '已关闭', value: 'closed', color: 'default' }
};

function statusLabel(status) {
  let statusOption = StatusType[status];

  return <Label color={statusOption?.color || 'secondary'}> {statusOption?.name || '未知'} </Label>;
}

export { StatusType };

export default function OrderTableRow({ item }) {
  return (
    <>
      <TableRow tabIndex={item.id}>
        <TableCell>{timestamp2string(item.created_at)}</TableCell>
        <TableCell>{item.user_id}</TableCell>
        <TableCell>{item.trade_no}</TableCell>
        <TableCell>{item.gateway_no}</TableCell>
        <TableCell>${item.amount}</TableCell>
        <TableCell>${item.fee}</TableCell>
        <TableCell>
          {item.order_amount} {item.order_currency}
        </TableCell>
        <TableCell>{item.quota}</TableCell>
        <TableCell>{statusLabel(item.status)}</TableCell>
      </TableRow>
    </>
  );
}

OrderTableRow.propTypes = {
  item: PropTypes.object
};
