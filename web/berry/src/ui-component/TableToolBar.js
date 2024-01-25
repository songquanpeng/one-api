import PropTypes from 'prop-types';

import Toolbar from '@mui/material/Toolbar';
import OutlinedInput from '@mui/material/OutlinedInput';
import InputAdornment from '@mui/material/InputAdornment';

import { useTheme } from '@mui/material/styles';
import { IconSearch } from '@tabler/icons-react';

// ----------------------------------------------------------------------

export default function TableToolBar({ filterName, handleFilterName, placeholder }) {
  const theme = useTheme();
  const grey500 = theme.palette.grey[500];

  return (
    <Toolbar
      sx={{
        height: 80,
        display: 'flex',
        justifyContent: 'space-between',
        p: (theme) => theme.spacing(0, 1, 0, 3)
      }}
    >
      <OutlinedInput
        id="keyword"
        sx={{
          minWidth: '100%'
        }}
        value={filterName}
        onChange={handleFilterName}
        placeholder={placeholder}
        startAdornment={
          <InputAdornment position="start">
            <IconSearch stroke={1.5} size="20px" color={grey500} />
          </InputAdornment>
        }
      />
    </Toolbar>
  );
}

TableToolBar.propTypes = {
  filterName: PropTypes.string,
  handleFilterName: PropTypes.func,
  placeholder: PropTypes.string
};
