import { closeSnackbar } from 'notistack';
import { IconX } from '@tabler/icons-react';
import { IconButton } from '@mui/material';
const action = (snackbarId) => (
  <>
    <IconButton
      onClick={() => {
        closeSnackbar(snackbarId);
      }}
    >
      <IconX stroke={1.5} size="1.25rem" />
    </IconButton>
  </>
);

export const snackbarConstants = {
  Common: {
    ERROR: {
      variant: 'error',
      autoHideDuration: 5000,
      preventDuplicate: true,
      action
    },
    WARNING: {
      variant: 'warning',
      autoHideDuration: 10000,
      preventDuplicate: true,
      action
    },
    SUCCESS: {
      variant: 'success',
      autoHideDuration: 1500,
      preventDuplicate: true,
      action
    },
    INFO: {
      variant: 'info',
      autoHideDuration: 3000,
      preventDuplicate: true,
      action
    },
    NOTICE: {
      variant: 'info',
      autoHideDuration: 20000,
      preventDuplicate: true,
      action
    },
    COPY: {
      variant: 'copy',
      persist: true,
      preventDuplicate: true,
      allowDownload: true,
      action
    }
  },
  Mobile: {
    anchorOrigin: { vertical: 'bottom', horizontal: 'center' }
  }
};
