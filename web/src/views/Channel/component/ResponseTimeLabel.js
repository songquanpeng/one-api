import PropTypes from 'prop-types';
import Label from 'ui-component/Label';
import Tooltip from '@mui/material/Tooltip';
import { timestamp2string } from 'utils/common';

const ResponseTimeLabel = ({ test_time, response_time, handle_action }) => {
  let color = 'default';
  let time = response_time / 1000;
  time = time.toFixed(2) + ' 秒';

  if (response_time === 0) {
    color = 'default';
  } else if (response_time <= 1000) {
    color = 'success';
  } else if (response_time <= 3000) {
    color = 'primary';
  } else if (response_time <= 5000) {
    color = 'secondary';
  } else {
    color = 'error';
  }
  let title = (
    <>
      点击测速
      <br />
      {test_time != 0 ? '上次测速时间：' + timestamp2string(test_time) : '未测试'}
    </>
  );

  return (
    <Tooltip title={title} placement="top" onClick={handle_action}>
      <Label color={color}> {response_time == 0 ? '未测试' : time} </Label>
    </Tooltip>
  );
};

ResponseTimeLabel.propTypes = {
  test_time: PropTypes.number,
  response_time: PropTypes.number,
  handle_action: PropTypes.func
};

export default ResponseTimeLabel;
