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
  Tooltip,
  Stack
} from '@mui/material';

import Label from 'ui-component/Label';
import TableSwitch from 'ui-component/Switch';
import { renderQuota, renderNumber } from 'utils/common';
import { IconDotsVertical, IconEdit, IconTrash, IconUser, IconBrandWechat, IconBrandGithub, IconMail } from '@tabler/icons-react';
import { useTheme } from '@mui/material/styles';

function renderRole(role) {
  switch (role) {
    case 1:
      return <Label color="default">普通用户</Label>;
    case 10:
      return <Label color="orange">管理员</Label>;
    case 100:
      return <Label color="success">超级管理员</Label>;
    default:
      return <Label color="error">未知身份</Label>;
  }
}

export default function UsersTableRow({ item, manageUser, handleOpenModal, setModalUserId }) {
  const theme = useTheme();
  const [open, setOpen] = useState(null);
  const [openDelete, setOpenDelete] = useState(false);
  const [statusSwitch, setStatusSwitch] = useState(item.status);

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

  const handleStatus = async () => {
    const switchVlue = statusSwitch === 1 ? 2 : 1;
    const { success } = await manageUser(item.username, 'status', switchVlue);
    if (success) {
      setStatusSwitch(switchVlue);
    }
  };

  const handleDelete = async () => {
    handleCloseMenu();
    await manageUser(item.username, 'delete', '');
  };

  return (
    <>
      <TableRow tabIndex={item.id}>
        <TableCell>{item.id}</TableCell>

        <TableCell>{item.username}</TableCell>

        <TableCell>
          <Label>{item.group}</Label>
        </TableCell>

        <TableCell>
          <Stack direction="row" spacing={0.5} alignItems="center" justifyContent="center">
            <Tooltip title={'剩余额度'} placement="top">
              <Label color={'primary'} variant="outlined">
                {' '}
                {renderQuota(item.quota)}{' '}
              </Label>
            </Tooltip>
            <Tooltip title={'已用额度'} placement="top">
              <Label color={'primary'} variant="outlined">
                {' '}
                {renderQuota(item.used_quota)}{' '}
              </Label>
            </Tooltip>
            <Tooltip title={'请求次数'} placement="top">
              <Label color={'primary'} variant="outlined">
                {' '}
                {renderNumber(item.request_count)}{' '}
              </Label>
            </Tooltip>
          </Stack>
        </TableCell>
        <TableCell>{renderRole(item.role)}</TableCell>
        <TableCell>
          <Stack direction="row" spacing={0.5} alignItems="center" justifyContent="center">
            <Tooltip title={item.wechat_id ? item.wechat_id : '未绑定'} placement="top">
              <IconBrandWechat color={item.wechat_id ? theme.palette.success.dark : theme.palette.grey[400]} />
            </Tooltip>
            <Tooltip title={item.github_id ? item.github_id : '未绑定'} placement="top">
              <IconBrandGithub color={item.github_id ? theme.palette.grey[900] : theme.palette.grey[400]} />
            </Tooltip>
            <Tooltip title={item.email ? item.email : '未绑定'} placement="top">
              <IconMail color={item.email ? theme.palette.grey[900] : theme.palette.grey[400]} />
            </Tooltip>
          </Stack>
        </TableCell>

        <TableCell>
          {' '}
          <TableSwitch id={`switch-${item.id}`} checked={statusSwitch === 1} onChange={handleStatus} />
        </TableCell>
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
          sx: { width: 140 }
        }}
      >
        {item.role !== 100 && (
          <MenuItem
            onClick={() => {
              handleCloseMenu();
              manageUser(item.username, 'role', item.role === 1 ? true : false);
            }}
          >
            <IconUser style={{ marginRight: '16px' }} />
            {item.role === 1 ? '设为管理员' : '取消管理员'}
          </MenuItem>
        )}

        <MenuItem
          onClick={() => {
            handleCloseMenu();
            handleOpenModal();
            setModalUserId(item.id);
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
        <DialogTitle>删除用户</DialogTitle>
        <DialogContent>
          <DialogContentText>是否删除用户 {item.name}？</DialogContentText>
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

UsersTableRow.propTypes = {
  item: PropTypes.object,
  manageUser: PropTypes.func,
  handleOpenModal: PropTypes.func,
  setModalUserId: PropTypes.func
};
