import PropTypes from 'prop-types';

import { useState } from 'react';
import {
  TableRow,
  TableCell,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  ButtonGroup,
  Popover,
  MenuItem,
  MenuList,
  Tooltip
} from '@mui/material';

import { timestamp2string, copy } from 'utils/common';
import Label from 'ui-component/Label';
import { ACTION_TYPE, CODE_TYPE, STATUS_TYPE } from '../type/Type';
import { IconCaretDownFilled, IconCopy, IconDownload, IconExternalLink } from '@tabler/icons-react';

function renderType(types, type) {
  const typeOption = types[type];
  if (typeOption) {
    return (
      <Label variant="filled" color={typeOption.color}>
        {' '}
        {typeOption.text}{' '}
      </Label>
    );
  } else {
    return (
      <Label variant="filled" color="error">
        {' '}
        未知{' '}
      </Label>
    );
  }
}
async function downloadImage(url, filename) {
  const response = await fetch(url);
  const blob = await response.blob();
  const blobUrl = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = blobUrl;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(blobUrl);
}

function TruncatedText(text) {
  const truncatedText = text.length > 30 ? text.substring(0, 100) + '...' : text;

  return (
    <Tooltip
      placement="top"
      title={text}
      onClick={() => {
        copy(text, '');
      }}
    >
      <span>{truncatedText}</span>
    </Tooltip>
  );
}

export default function LogTableRow({ item, userIsAdmin }) {
  const [open, setOpen] = useState(false);
  const [menuOpen, setMenuOpen] = useState(null);
  const handleClickOpen = () => {
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
  };

  const handleOpenMenu = (event) => {
    setMenuOpen(event.currentTarget);
  };

  const handleCloseMenu = () => {
    setMenuOpen(null);
  };

  return (
    <>
      <TableRow tabIndex={item.id}>
        <TableCell>{item.mj_id}</TableCell>
        <TableCell>{timestamp2string(item.submit_time / 1000)}</TableCell>

        {userIsAdmin && <TableCell>{item.channel_id || ''}</TableCell>}
        {userIsAdmin && <TableCell>{item.user_id || ''}</TableCell>}

        <TableCell>{renderType(ACTION_TYPE, item.action)}</TableCell>
        {userIsAdmin && <TableCell>{renderType(CODE_TYPE, item.code)}</TableCell>}
        {userIsAdmin && <TableCell>{renderType(STATUS_TYPE, item.status)}</TableCell>}
        <TableCell>{item.progress}</TableCell>
        <TableCell>
          {item.image_url == '' ? (
            '无'
          ) : (
            <ButtonGroup size="small" aria-label="split button">
              <Button color="primary" onClick={handleClickOpen}>
                显示
              </Button>
              <Button onClick={handleOpenMenu}>
                <IconCaretDownFilled size={'16px'} />
              </Button>
            </ButtonGroup>
          )}
        </TableCell>
        <TableCell>{TruncatedText(item.prompt)}</TableCell>
        <TableCell>{TruncatedText(item.prompt_en)}</TableCell>
        <TableCell>{TruncatedText(item.fail_reason)}</TableCell>
      </TableRow>
      <Dialog open={open} onClose={handleClose}>
        <DialogContent>
          <img src={item.image_url} alt="item" style={{ maxWidth: '100%', maxHeight: '100%' }} />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">
            关闭
          </Button>
        </DialogActions>
      </Dialog>

      <Popover
        open={!!menuOpen}
        anchorEl={menuOpen}
        onClose={handleCloseMenu}
        anchorOrigin={{ vertical: 'top', horizontal: 'left' }}
        transformOrigin={{ vertical: 'top', horizontal: 'right' }}
        PaperProps={{
          sx: { width: 140 }
        }}
      >
        <MenuList>
          <MenuItem
            onClick={() => {
              handleCloseMenu();
              copy(item.image_url, '图片地址');
            }}
          >
            <IconCopy style={{ marginRight: '16px' }} />
            复制地址
          </MenuItem>

          <MenuItem
            onClick={async () => {
              handleCloseMenu();
              await downloadImage(item.image_url, item.mj_id + '.png');
            }}
          >
            <IconDownload style={{ marginRight: '16px' }} /> 下载图片{' '}
          </MenuItem>
          <MenuItem
            onClick={() => {
              handleCloseMenu();
            }}
          >
            <IconExternalLink style={{ marginRight: '16px' }} /> 新窗口打开{' '}
          </MenuItem>
        </MenuList>
      </Popover>
    </>
  );
}

LogTableRow.propTypes = {
  item: PropTypes.object,
  userIsAdmin: PropTypes.bool
};
