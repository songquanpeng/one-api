import PropTypes from 'prop-types';

import { TableRow, TableCell } from '@mui/material';

import Label from 'ui-component/Label';
import { copy } from 'utils/common';

export default function PricesTableRow({ item }) {
  return (
    <>
      <TableRow tabIndex={item.model}>
        <TableCell>
          <Label
            variant="outlined"
            color="primary"
            key={item.model}
            onClick={() => {
              copy(item.model, '模型名称');
            }}
          >
            {item.model}
          </Label>
        </TableCell>

        <TableCell>{item.type}</TableCell>
        <TableCell>{item.channel_type}</TableCell>

        <TableCell>{item.input}</TableCell>
        <TableCell>{item.output}</TableCell>
      </TableRow>
    </>
  );
}

PricesTableRow.propTypes = {
  item: PropTypes.object,
  userModelList: PropTypes.object,
  ownedby: PropTypes.array
};
