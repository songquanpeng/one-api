import PropTypes from 'prop-types';
import { useState } from 'react';

import {
  TableRow,
  TableCell,
  Popover,
  MenuItem,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  Button,
  TextField
} from '@mui/material';

import { IconDotsVertical, IconEdit, IconTrash } from '@tabler/icons-react';
import { timestamp2string, showError } from 'utils/common';
import { PaymentType } from '../type/Config';
import TableSwitch from 'ui-component/Switch';

export default function PaymentTableRow({ item, managePayment, handleOpenModal, setModalPaymentId }) {
  const [open, setOpen] = useState(null);
  const [openDelete, setOpenDelete] = useState(false);
  const [sortValve, setSort] = useState(item.sort);

  const handleCloseMenu = () => {
    setOpen(null);
  };
  const handleOpenMenu = (event) => {
    setOpen(event.currentTarget);
  };

  const handleDeleteOpen = () => {
    handleCloseMenu();
    setOpenDelete(true);
  };

  const handleDeleteClose = () => {
    setOpenDelete(false);
  };

  const handleDelete = async () => {
    handleCloseMenu();
    await managePayment(item.id, 'delete', '');
  };

  const handleSort = async (event) => {
    const currentValue = parseInt(event.target.value);
    if (isNaN(currentValue) || currentValue === sortValve) {
      return;
    }

    if (currentValue < 0) {
      showError('排序不能小于 0');
      return;
    }

    await managePayment(item.id, 'sort', currentValue);
    setSort(currentValue);
  };

  return (
    <>
      <TableRow tabIndex={item.id}>
        <TableCell>{item.id}</TableCell>
        <TableCell>{item.uuid}</TableCell>
        <TableCell>{item.name}</TableCell>
        <TableCell>{PaymentType?.[item.type] || '未知'}</TableCell>
        <TableCell>
          <img src={item.icon} alt="icon" style={{ width: '24px', height: '24px' }} />
        </TableCell>
        <TableCell>{item.fixed_fee}</TableCell>
        <TableCell>{item.percent_fee}</TableCell>
        <TableCell>
          <TextField
            id={`sort-${item.id}`}
            onBlur={handleSort}
            type="number"
            label="排序"
            variant="standard"
            defaultValue={item.sort}
            inputProps={{ min: '0' }}
          />
        </TableCell>
        <TableCell>
          <TableSwitch
            id={`switch-${item.id}`}
            checked={item.enable}
            onChange={() => {
              managePayment(item.id, 'status', !item.enable);
            }}
          />
        </TableCell>
        <TableCell>{timestamp2string(item.created_at)}</TableCell>
        <TableCell>
          <IconButton onClick={handleOpenMenu} sx={{ color: 'rgb(99, 115, 129)' }}>
            <IconDotsVertical />
          </IconButton>
        </TableCell>
      </TableRow>

      <Popover
        open={!!open}
        anchorEl={open}
        onClose={handleCloseMenu}
        anchorOrigin={{ vertical: 'top', horizontal: 'left' }}
        transformOrigin={{ vertical: 'top', horizontal: 'right' }}
        PaperProps={{
          sx: { minWidth: 140 }
        }}
      >
        <MenuItem
          onClick={() => {
            handleCloseMenu();
            handleOpenModal();
            setModalPaymentId(item.id);
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
        <DialogTitle>删除通道</DialogTitle>
        <DialogContent>
          <DialogContentText>是否删除通道 {item.name}？</DialogContentText>
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

PaymentTableRow.propTypes = {
  item: PropTypes.object,
  managePayment: PropTypes.func,
  handleOpenModal: PropTypes.func,
  setModalPaymentId: PropTypes.func
};
