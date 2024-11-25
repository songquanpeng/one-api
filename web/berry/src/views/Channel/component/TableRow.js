import PropTypes from "prop-types";
import { useState } from "react";

import { showInfo, showError, renderNumber } from "utils/common";
import { API } from "utils/api";
import { CHANNEL_OPTIONS } from "constants/ChannelConstants";

import {
  Popover,
  TableRow,
  MenuItem,
  TableCell,
  IconButton,
  TextField,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Tooltip,
  Button,
} from "@mui/material";

import Label from "ui-component/Label";
import TableSwitch from "ui-component/Switch";

import ResponseTimeLabel from "./ResponseTimeLabel";
import GroupLabel from "./GroupLabel";
import NameLabel from "./NameLabel";

import { IconDotsVertical, IconEdit, IconTrash } from "@tabler/icons-react";

export default function ChannelTableRow({
  item,
  manageChannel,
  handleOpenModal,
  setModalChannelId,
}) {
  const [open, setOpen] = useState(null);
  const [openDelete, setOpenDelete] = useState(false);
  const [statusSwitch, setStatusSwitch] = useState(item.status);
  const [priorityValve, setPriority] = useState(item.priority);
  const [responseTimeData, setResponseTimeData] = useState({
    test_time: item.test_time,
    response_time: item.response_time,
  });
  const [itemBalance, setItemBalance] = useState(item.balance);

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
    const { success } = await manageChannel(item.id, "status", switchVlue);
    if (success) {
      setStatusSwitch(switchVlue);
    }
  };

  const handlePriority = async (event) => {
    const currentValue = parseInt(event.target.value);
    if (isNaN(currentValue) || currentValue === priorityValve) {
      return;
    }

    if (currentValue < 0) {
      showError("优先级不能小于 0");
      return;
    }

    await manageChannel(item.id, "priority", currentValue);
    setPriority(currentValue);
  };

  const handleResponseTime = async () => {
    const { success, time } = await manageChannel(item.id, "test", "");
    if (success) {
      setResponseTimeData({
        test_time: Date.now() / 1000,
        response_time: time * 1000,
      });
      showInfo(`渠道 ${item.name} 测试成功，耗时 ${time.toFixed(2)} 秒。`);
    }
  };

  const updateChannelBalance = async () => {
    const res = await API.get(`/api/channel/update_balance/${item.id}`);
    const { success, message, balance } = res.data;
    if (success) {
      setItemBalance(balance);

      showInfo(`余额更新成功！`);
    } else {
      showError(message);
    }
  };

  const handleDelete = async () => {
    handleCloseMenu();
    await manageChannel(item.id, "delete", "");
  };

  return (
    <>
      <TableRow tabIndex={item.id}>
        <TableCell>{item.id}</TableCell>

        <TableCell>
          <NameLabel name={item.name} models={item.models} />
        </TableCell>

        <TableCell>
          <GroupLabel group={item.group} />
        </TableCell>

        <TableCell>
          {!CHANNEL_OPTIONS[item.type] ? (
            <Label color="error" variant="outlined">
              未知
            </Label>
          ) : (
            <Label color={CHANNEL_OPTIONS[item.type].color} variant="outlined">
              {CHANNEL_OPTIONS[item.type].text}
            </Label>
          )}
        </TableCell>

        <TableCell>
          <Tooltip
            title={(() => {
              switch (statusSwitch) {
                case 1:
                  return "已启用";
                case 2:
                  return "本渠道被手动禁用";
                case 3:
                  return "本渠道被程序自动禁用";
                default:
                  return "未知";
              }
            })()}
            placement="top"
          >
            <TableSwitch
              id={`switch-${item.id}`}
              checked={statusSwitch === 1}
              onChange={handleStatus}
            />
          </Tooltip>
        </TableCell>

        <TableCell>
          <ResponseTimeLabel
            test_time={responseTimeData.test_time}
            response_time={responseTimeData.response_time}
            handle_action={handleResponseTime}
          />
        </TableCell>
        <TableCell>{renderNumber(item.used_quota)}</TableCell>
        <TableCell>
          <Tooltip
            title={"点击更新余额"}
            placement="top"
            onClick={updateChannelBalance}
          >
            {renderBalance(item.type, itemBalance)}
          </Tooltip>
        </TableCell>
        <TableCell>
          <TextField
            id={`priority-${item.id}`}
            onBlur={handlePriority}
            type="number"
            label="优先级"
            variant="standard"
            defaultValue={item.priority}
            inputProps={{ min: "0" }}
            sx={{ width: 80 }}
          />
        </TableCell>

        <TableCell>
          <IconButton
            onClick={handleOpenMenu}
            sx={{ color: "rgb(99, 115, 129)" }}
          >
            <IconDotsVertical />
          </IconButton>
        </TableCell>
      </TableRow>

      <Popover
        open={!!open}
        anchorEl={open}
        onClose={handleCloseMenu}
        anchorOrigin={{ vertical: "top", horizontal: "left" }}
        transformOrigin={{ vertical: "top", horizontal: "right" }}
        PaperProps={{
          sx: { width: 140 },
        }}
      >
        <MenuItem
          onClick={() => {
            handleCloseMenu();
            handleOpenModal();
            setModalChannelId(item.id);
          }}
        >
          <IconEdit style={{ marginRight: "16px" }} />
          编辑
        </MenuItem>
        <MenuItem onClick={handleDeleteOpen} sx={{ color: "error.main" }}>
          <IconTrash style={{ marginRight: "16px" }} />
          删除
        </MenuItem>
      </Popover>

      <Dialog open={openDelete} onClose={handleDeleteClose}>
        <DialogTitle>删除渠道</DialogTitle>
        <DialogContent>
          <DialogContentText>是否删除渠道 {item.name}？</DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeleteClose}>关闭</Button>
          <Button onClick={handleDelete} sx={{ color: "error.main" }} autoFocus>
            删除
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}

ChannelTableRow.propTypes = {
  item: PropTypes.object,
  manageChannel: PropTypes.func,
  handleOpenModal: PropTypes.func,
  setModalChannelId: PropTypes.func,
};

function renderBalance(type, balance) {
  switch (type) {
    case 1: // OpenAI
      return <span>${balance.toFixed(2)}</span>;
    case 4: // CloseAI
      return <span>¥{balance.toFixed(2)}</span>;
    case 8: // 自定义
      return <span>${balance.toFixed(2)}</span>;
    case 5: // OpenAI-SB
      return <span>¥{(balance / 10000).toFixed(2)}</span>;
    case 10: // AI Proxy
      return <span>{renderNumber(balance)}</span>;
    case 12: // API2GPT
      return <span>¥{balance.toFixed(2)}</span>;
    case 13: // AIGC2D
      return <span>{renderNumber(balance)}</span>;
    case 36: // DeepSeek
      return <span>¥{balance.toFixed(2)}</span>;
    case 44: // SiliconFlow
      return <span>¥{balance.toFixed(2)}</span>;
    default:
      return <span>不支持</span>;
  }
}
