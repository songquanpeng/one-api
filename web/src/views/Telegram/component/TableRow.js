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
  Button,
  Stack
} from '@mui/material';

import { IconDotsVertical, IconEdit, IconTrash } from '@tabler/icons-react';

export default function TelegramTableRow({ item, manageAction, handleOpenModal, setModalId }) {
  const [open, setOpen] = useState(null);
  const [openDelete, setOpenDelete] = useState(false);

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
    handleCloseMenu();
    await manageAction(item.id, 'delete');
  };

  return (
    <>
      <TableRow tabIndex={item.id}>
        <TableCell>{item.id}</TableCell>

        <TableCell>{item.command}</TableCell>

        <TableCell>{item.description}</TableCell>

        <TableCell>{item.parse_mode}</TableCell>
        <TableCell>{item.reply_message}</TableCell>
        <TableCell>
          <Stack direction="row" spacing={1}>
            <IconButton onClick={handleOpenMenu} sx={{ color: 'rgb(99, 115, 129)' }}>
              <IconDotsVertical />
            </IconButton>
          </Stack>
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
            setModalId(item.id);
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
        <DialogTitle>删除菜单</DialogTitle>
        <DialogContent>
          <DialogContentText>是否删除菜单 {item.name}？</DialogContentText>
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

TelegramTableRow.propTypes = {
  item: PropTypes.object,
  manageAction: PropTypes.func,
  handleOpenModal: PropTypes.func,
  setModalId: PropTypes.func
};
