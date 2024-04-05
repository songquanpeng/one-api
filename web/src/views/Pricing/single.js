import PropTypes from 'prop-types';
import { useState, useEffect, useMemo, useCallback } from 'react';
import { GridRowModes, DataGrid, GridToolbarContainer, GridActionsCellItem } from '@mui/x-data-grid';
import { Box, Button, Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/DeleteOutlined';
import SaveIcon from '@mui/icons-material/Save';
import CancelIcon from '@mui/icons-material/Close';
import { showError, showSuccess } from 'utils/common';
import { API } from 'utils/api';
import { ValueFormatter, priceType } from './component/util';

function validation(row, rows) {
  if (row.model === '') {
    return '模型名称不能为空';
  }

  // 判断 type 是否是 等于 tokens || times
  if (row.type !== 'tokens' && row.type !== 'times') {
    return '类型只能是tokens或times';
  }

  if (row.channel_type <= 0) {
    return '所属渠道类型错误';
  }

  // 判断 model是否是唯一值
  if (rows.filter((r) => r.model === row.model && (row.isNew || r.id !== row.id)).length > 0) {
    return '模型名称不能重复';
  }

  if (row.input === '' || row.input < 0) {
    return '输入倍率必须大于等于0';
  }
  if (row.output === '' || row.output < 0) {
    return '输出倍率必须大于等于0';
  }
  return false;
}

function randomId() {
  return Math.random().toString(36).substr(2, 9);
}

function EditToolbar({ setRows, setRowModesModel }) {
  const handleClick = () => {
    const id = randomId();
    setRows((oldRows) => [{ id, model: '', type: 'tokens', channel_type: 1, input: 0, output: 0, isNew: true }, ...oldRows]);
    setRowModesModel((oldModel) => ({
      [id]: { mode: GridRowModes.Edit, fieldToFocus: 'name' },
      ...oldModel
    }));
  };

  return (
    <GridToolbarContainer>
      <Button color="primary" startIcon={<AddIcon />} onClick={handleClick}>
        新增
      </Button>
    </GridToolbarContainer>
  );
}

EditToolbar.propTypes = {
  setRows: PropTypes.func.isRequired,
  setRowModesModel: PropTypes.func.isRequired
};

const Single = ({ ownedby, prices, reloadData }) => {
  const [rows, setRows] = useState([]);
  const [rowModesModel, setRowModesModel] = useState({});
  const [selectedRow, setSelectedRow] = useState(null);

  const addOrUpdatePirces = useCallback(
    async (newRow, oldRow, reject, resolve) => {
      try {
        let res;
        if (oldRow.model == '') {
          res = await API.post('/api/prices/single', newRow);
        } else {
          let modelEncode = encodeURIComponent(oldRow.model);
          res = await API.put('/api/prices/single/' + modelEncode, newRow);
        }
        const { success, message } = res.data;
        if (success) {
          showSuccess('保存成功');
          resolve(newRow);
          reloadData();
        } else {
          reject(new Error(message));
        }
      } catch (error) {
        reject(new Error(error));
      }
    },
    [reloadData]
  );

  const handleEditClick = useCallback(
    (id) => () => {
      setRowModesModel({ ...rowModesModel, [id]: { mode: GridRowModes.Edit } });
    },
    [rowModesModel]
  );

  const handleSaveClick = useCallback(
    (id) => () => {
      setRowModesModel({ ...rowModesModel, [id]: { mode: GridRowModes.View } });
    },
    [rowModesModel]
  );

  const handleDeleteClick = useCallback(
    (id) => () => {
      setSelectedRow(rows.find((row) => row.id === id));
    },
    [rows]
  );

  const handleClose = () => {
    setSelectedRow(null);
  };

  const handleConfirmDelete = async () => {
    // 执行删除操作
    await deletePirces(selectedRow.model);
    setSelectedRow(null);
  };

  const handleCancelClick = useCallback(
    (id) => () => {
      setRowModesModel({
        ...rowModesModel,
        [id]: { mode: GridRowModes.View, ignoreModifications: true }
      });

      const editedRow = rows.find((row) => row.id === id);
      if (editedRow.isNew) {
        setRows(rows.filter((row) => row.id !== id));
      }
    },
    [rowModesModel, rows]
  );

  const processRowUpdate = useCallback(
    (newRow, oldRows) =>
      new Promise((resolve, reject) => {
        if (
          !newRow.isNew &&
          newRow.model === oldRows.model &&
          newRow.input === oldRows.input &&
          newRow.output === oldRows.output &&
          newRow.type === oldRows.type &&
          newRow.channel_type === oldRows.channel_type
        ) {
          return resolve(oldRows);
        }
        const updatedRow = { ...newRow, isNew: false };
        const error = validation(updatedRow, rows);
        if (error) {
          return reject(new Error(error));
        }

        const response = addOrUpdatePirces(updatedRow, oldRows, reject, resolve);
        return response;
      }),
    [rows, addOrUpdatePirces]
  );

  const handleProcessRowUpdateError = useCallback((error) => {
    showError(error.message);
  }, []);

  const handleRowModesModelChange = (newRowModesModel) => {
    setRowModesModel(newRowModesModel);
  };

  const modelRatioColumns = useMemo(
    () => [
      {
        field: 'model',
        sortable: true,
        headerName: '模型名称',
        minWidth: 220,
        flex: 1,
        editable: true,
        hideable: false
      },
      {
        field: 'type',
        sortable: true,
        headerName: '类型',
        flex: 1,
        minWidth: 220,
        type: 'singleSelect',
        valueOptions: priceType,
        editable: true,
        hideable: false
      },
      {
        field: 'channel_type',
        sortable: true,
        headerName: '供应商',
        flex: 1,
        minWidth: 220,
        type: 'singleSelect',
        valueOptions: ownedby,
        editable: true,
        hideable: false
      },
      {
        field: 'input',
        sortable: false,
        headerName: '输入倍率',
        flex: 0.8,
        minWidth: 150,
        type: 'number',
        editable: true,
        valueFormatter: (params) => ValueFormatter(params.value),
        hideable: false
      },
      {
        field: 'output',
        sortable: false,
        headerName: '输出倍率',
        flex: 0.8,
        minWidth: 150,
        type: 'number',
        editable: true,
        valueFormatter: (params) => ValueFormatter(params.value),
        hideable: false
      },
      {
        field: 'actions',
        type: 'actions',
        headerName: '操作',
        flex: 0.5,
        minWidth: 100,
        // width: 100,
        cellClassName: 'actions',
        hideable: false,
        getActions: ({ id }) => {
          const isInEditMode = rowModesModel[id]?.mode === GridRowModes.Edit;

          if (isInEditMode) {
            return [
              <GridActionsCellItem
                icon={<SaveIcon />}
                key={'Save-' + id}
                label="Save"
                sx={{
                  color: 'primary.main'
                }}
                onClick={handleSaveClick(id)}
              />,
              <GridActionsCellItem
                icon={<CancelIcon />}
                key={'Cancel-' + id}
                label="Cancel"
                className="textPrimary"
                onClick={handleCancelClick(id)}
                color="inherit"
              />
            ];
          }

          return [
            <GridActionsCellItem
              key={'Edit-' + id}
              icon={<EditIcon />}
              label="Edit"
              className="textPrimary"
              onClick={handleEditClick(id)}
              color="inherit"
            />,
            <GridActionsCellItem
              key={'Delete-' + id}
              icon={<DeleteIcon />}
              label="Delete"
              onClick={handleDeleteClick(id)}
              color="inherit"
            />
          ];
        }
      }
    ],
    [handleCancelClick, handleDeleteClick, handleEditClick, handleSaveClick, rowModesModel, ownedby]
  );

  const deletePirces = async (modelName) => {
    try {
      let modelEncode = encodeURIComponent(modelName);
      const res = await API.delete('/api/prices/single/' + modelEncode);
      const { success, message } = res.data;
      if (success) {
        showSuccess('保存成功');
        await reloadData();
      } else {
        showError(message);
      }
    } catch (error) {
      console.error(error);
    }
  };

  useEffect(() => {
    let modelRatioList = [];
    let id = 0;
    for (let key in prices) {
      modelRatioList.push({ id: id++, ...prices[key] });
    }
    setRows(modelRatioList);
  }, [prices]);

  return (
    <Box
      sx={{
        width: '100%',
        '& .actions': {
          color: 'text.secondary'
        },
        '& .textPrimary': {
          color: 'text.primary'
        }
      }}
    >
      <DataGrid
        autoHeight
        autosizeOnMount
        rows={rows}
        columns={modelRatioColumns}
        editMode="row"
        hideFooter
        disableRowSelectionOnClick
        rowModesModel={rowModesModel}
        onRowModesModelChange={handleRowModesModelChange}
        processRowUpdate={processRowUpdate}
        onProcessRowUpdateError={handleProcessRowUpdateError}
        // onCellDoubleClick={(params, event) => {
        //   event.defaultMuiPrevented = true;
        // }}
        onRowEditStop={(params, event) => {
          if (params.reason === 'rowFocusOut') {
            event.defaultMuiPrevented = true;
          }
        }}
        slots={{
          toolbar: EditToolbar
        }}
        slotProps={{
          toolbar: { setRows, setRowModesModel }
        }}
      />

      <Dialog
        maxWidth="xs"
        // TransitionProps={{ onEntered: handleEntered }}
        open={!!selectedRow}
      >
        <DialogTitle>确定删除?</DialogTitle>
        <DialogContent dividers>{`确定删除 ${selectedRow?.model} 吗？`}</DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>取消</Button>
          <Button onClick={handleConfirmDelete}>删除</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Single;

Single.propTypes = {
  prices: PropTypes.array,
  ownedby: PropTypes.array,
  reloadData: PropTypes.func
};
