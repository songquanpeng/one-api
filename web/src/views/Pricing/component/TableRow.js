import PropTypes from 'prop-types';
import { useState } from 'react';

import {
  Popover,
  TableRow,
  MenuItem,
  TableCell,
  IconButton,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Collapse,
  Grid,
  Box,
  Typography,
  Button
} from '@mui/material';

import { IconDotsVertical, IconEdit, IconTrash } from '@tabler/icons-react';
import { ValueFormatter, priceType } from './util';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import Label from 'ui-component/Label';
import { copy } from 'utils/common';

export default function PricesTableRow({ item, managePrices, handleOpenModal, setModalPricesItem, ownedby }) {
  const [open, setOpen] = useState(null);
  const [openRow, setOpenRow] = useState(false);
  const [openDelete, setOpenDelete] = useState(false);
  const type_label = priceType.find((pt) => pt.value === item.type);
  const channel_label = ownedby.find((ob) => ob.value === item.channel_type);
  const handleDeleteOpen = () => {
    handleCloseMenu();
    setOpenDelete(true);
  };

  const handleDeleteClose = () => {
    setOpenDelete(false);
  };

  const handleOpenMenu = (event) => {
    setOpen(event.currentTarget);
  };

  const handleCloseMenu = () => {
    setOpen(null);
  };

  const handleDelete = async () => {
    handleDeleteClose();
    await managePrices(item, 'delete', '');
  };

  return (
    <>
      <TableRow tabIndex={item.id} onClick={() => setOpenRow(!openRow)}>
        <TableCell>
          <IconButton aria-label="expand row" size="small" onClick={() => setOpenRow(!openRow)}>
            {openRow ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
          </IconButton>
        </TableCell>

        <TableCell>{type_label?.label}</TableCell>

        <TableCell>{channel_label?.label}</TableCell>
        <TableCell>{ValueFormatter(item.input)}</TableCell>
        <TableCell>{ValueFormatter(item.output)}</TableCell>
        <TableCell>{item.models.length}</TableCell>

        <TableCell onClick={(event) => event.stopPropagation()}>
          <IconButton onClick={handleOpenMenu} sx={{ color: 'rgb(99, 115, 129)' }}>
            <IconDotsVertical />
          </IconButton>
        </TableCell>
      </TableRow>

      <TableRow>
        <TableCell style={{ paddingBottom: 0, paddingTop: 0, textAlign: 'left' }} colSpan={10}>
          <Collapse in={openRow} timeout="auto" unmountOnExit>
            <Grid container spacing={1}>
              <Grid item xs={12}>
                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: '10px', margin: 1 }}>
                  <Typography variant="h6" gutterBottom component="div">
                    可用模型:
                  </Typography>
                  {item.models.map((model) => (
                    <Label
                      variant="outlined"
                      color="primary"
                      key={model}
                      onClick={() => {
                        copy(model, '模型名称');
                      }}
                    >
                      {model}
                    </Label>
                  ))}
                </Box>
              </Grid>
            </Grid>
          </Collapse>
        </TableCell>
      </TableRow>
      <Popover
        open={!!open}
        anchorEl={open}
        onClose={handleCloseMenu}
        anchorOrigin={{ vertical: 'top', horizontal: 'left' }}
        transformOrigin={{ vertical: 'top', horizontal: 'right' }}
        PaperProps={{
          sx: { width: 140 }
        }}
      >
        <MenuItem
          onClick={() => {
            handleCloseMenu();
            handleOpenModal();
            setModalPricesItem(item);
          }}
        >
          <IconEdit style={{ marginRight: '16px' }} />
          编辑
        </MenuItem>
        <MenuItem onClick={handleDeleteOpen} sx={{ color: 'error.main' }}>
          <IconTrash style={{ marginRight: '16px' }} />
          删除
        </MenuItem>
      </Popover>

      <Dialog open={openDelete} onClose={handleDeleteClose}>
        <DialogTitle>删除价格组</DialogTitle>
        <DialogContent>
          <DialogContentText>是否删除价格组？</DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeleteClose}>关闭</Button>
          <Button onClick={handleDelete} sx={{ color: 'error.main' }} autoFocus>
            删除
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}

PricesTableRow.propTypes = {
  item: PropTypes.object,
  managePrices: PropTypes.func,
  handleOpenModal: PropTypes.func,
  setModalPricesItem: PropTypes.func,
  priceType: PropTypes.array,
  ownedby: PropTypes.array
};
