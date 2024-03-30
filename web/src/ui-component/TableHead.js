import PropTypes from 'prop-types';
import { TableCell, TableHead, TableRow, TableSortLabel, Tooltip, IconButton, Typography, Box } from '@mui/material';
import HelpOutlineIcon from '@mui/icons-material/HelpOutline';

const KeywordTableHead = ({ order, orderBy, headLabel, onRequestSort }) => {
  const onSort = (property) => (event) => {
    onRequestSort(event, property);
  };

  const label = (cell) => {
    if (cell.tooltip) {
      return (
        <Box display="flex" alignItems="center">
          <Typography variant="body1">{cell.label}</Typography>
          <Tooltip title={cell.tooltip} placement="bottom-start" enterDelay={300}>
            <IconButton size="small">
              <HelpOutlineIcon fontSize="inherit" />
            </IconButton>
          </Tooltip>
        </Box>
      );
    } else {
      return cell.label;
    }
  };

  return (
    <TableHead>
      <TableRow>
        {headLabel.map((headCell) =>
          headCell.hide && headCell.hide === true ? null : (
            <TableCell
              key={headCell.id}
              align={headCell.align || 'left'}
              // sortDirection={orderBy === headCell.id ? order : false}
              sx={{ width: headCell.width, minWidth: headCell.minWidth }}
            >
              {headCell.disableSort ? (
                label(headCell)
              ) : (
                <TableSortLabel
                  hideSortIcon
                  active={orderBy === headCell.id}
                  direction={orderBy === headCell.id ? order : 'asc'}
                  onClick={onSort(headCell.id)}
                >
                  {label(headCell)}
                </TableSortLabel>
              )}
            </TableCell>
          )
        )}
      </TableRow>
    </TableHead>
  );
};

export default KeywordTableHead;

KeywordTableHead.propTypes = {
  order: PropTypes.oneOf(['asc', 'desc']),
  orderBy: PropTypes.string,
  onRequestSort: PropTypes.func,
  headLabel: PropTypes.array
};
