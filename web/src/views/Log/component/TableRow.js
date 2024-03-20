import PropTypes from 'prop-types';

import { TableRow, TableCell } from '@mui/material';

import { timestamp2string, renderQuota } from 'utils/common';
import Label from 'ui-component/Label';
import LogType from '../type/LogType';

function renderType(type) {
  const typeOption = LogType[type];
  if (typeOption) {
    return (
      <Label variant="filled" color={typeOption.color}>
        {' '}
        {typeOption.text}{' '}
      </Label>
    );
  } else {
    return (
      <Label variant="filled" color="error">
        {' '}
        未知{' '}
      </Label>
    );
  }
}

function requestTimeLabelOptions(request_time) {
  let color = 'error';
  if (request_time === 0) {
    color = 'default';
  } else if (request_time <= 1000) {
    color = 'success';
  } else if (request_time <= 3000) {
    color = 'primary';
  } else if (request_time <= 5000) {
    color = 'secondary';
  }

  return color;
}

export default function LogTableRow({ item, userIsAdmin }) {
  let request_time = item.request_time / 1000;
  request_time = request_time.toFixed(2) + ' 秒';

  return (
    <>
      <TableRow tabIndex={item.id}>
        <TableCell>{timestamp2string(item.created_at)}</TableCell>

        {userIsAdmin && <TableCell>{item.channel || ''}</TableCell>}
        {userIsAdmin && (
          <TableCell>
            <Label color="default" variant="outlined">
              {item.username}
            </Label>
          </TableCell>
        )}
        <TableCell>
          {item.token_name && (
            <Label color="default" variant="soft">
              {item.token_name}
            </Label>
          )}
        </TableCell>
        <TableCell>{renderType(item.type)}</TableCell>
        <TableCell>
          {item.model_name && (
            <Label color="primary" variant="outlined">
              {item.model_name}
            </Label>
          )}
        </TableCell>
        <TableCell>
          {' '}
          <Label color={requestTimeLabelOptions(item.request_time)}> {item.request_time == 0 ? '无' : request_time} </Label>
        </TableCell>
        <TableCell>{item.prompt_tokens || '0'}</TableCell>
        <TableCell>{item.completion_tokens || '0'}</TableCell>
        <TableCell>{item.quota ? renderQuota(item.quota, 6) : '0'}</TableCell>
        <TableCell>{item.content}</TableCell>
      </TableRow>
    </>
  );
}

LogTableRow.propTypes = {
  item: PropTypes.object,
  userIsAdmin: PropTypes.bool
};
