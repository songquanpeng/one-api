export const snackbarConstants = {
  Common: {
    ERROR: {
      variant: 'error',
      autoHideDuration: 5000,
      preventDuplicate: true
    },
    WARNING: {
      variant: 'warning',
      autoHideDuration: 10000,
      preventDuplicate: true
    },
    SUCCESS: {
      variant: 'success',
      autoHideDuration: 1500,
      preventDuplicate: true
    },
    INFO: {
      variant: 'info',
      autoHideDuration: 3000,
      preventDuplicate: true
    },
    NOTICE: {
      variant: 'info',
      autoHideDuration: 20000,
      preventDuplicate: true
    },
    COPY: {
      variant: 'copy',
      persist: true,
      preventDuplicate: true,
      allowDownload: true
    }
  },
  Mobile: {
    anchorOrigin: { vertical: 'bottom', horizontal: 'center' }
  }
};
