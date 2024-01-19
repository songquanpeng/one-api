import PropTypes from 'prop-types';
import { forwardRef } from 'react';
import { SnackbarContent, useSnackbar } from 'notistack';
import { Alert } from '@mui/material';

const CopySnackbar = forwardRef((props, ref) => {
  const { closeSnackbar } = useSnackbar();

  return (
    <SnackbarContent ref={ref}>
      <Alert
        severity="info"
        variant="filled"
        sx={{ width: '100%', whiteSpace: 'normal', overflowWrap: 'break-word' }}
        onClose={() => {
          closeSnackbar(props.id);
        }}
      >
        {props.message}
      </Alert>
    </SnackbarContent>
  );
});

CopySnackbar.displayName = 'ReportComplete';

CopySnackbar.propTypes = {
  props: PropTypes.object,
  id: PropTypes.number.isRequired, // 添加 id 的验证
  message: PropTypes.any.isRequired // 添加 message 的验证
};

export default CopySnackbar;
