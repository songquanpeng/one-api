import PropTypes from 'prop-types';
import { useState, useEffect, useMemo, useCallback } from 'react';
import { GridRowModes, DataGrid, GridToolbarContainer, GridActionsCellItem } from '@mui/x-data-grid';
import { Box, Button } from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/DeleteOutlined';
import SaveIcon from '@mui/icons-material/Save';
import CancelIcon from '@mui/icons-material/Close';
import { showError } from 'utils/common';

function validation(row, rows) {
  if (row.model === '') {
    return '模型名称不能为空';
  }

  // 判断 model是否是唯一值
  if (rows.filter((r) => r.model === row.model && (row.isNew || r.id !== row.id)).length > 0) {
    return '模型名称不能重复';
  }

  if (row.input === '' || row.input < 0) {
    return '输入倍率必须大于等于0';
  }
  if (row.complete === '' || row.complete < 0) {
    return '完成倍率必须大于等于0';
  }
  return false;
}

function randomId() {
  return Math.random().toString(36).substr(2, 9);
}

function EditToolbar({ setRows, setRowModesModel }) {
  const handleClick = () => {
    const id = randomId();
    setRows((oldRows) => [{ id, model: '', input: 0, complete: 0, isNew: true }, ...oldRows]);
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

const ModelRationDataGrid = ({ ratio, onChange }) => {
  const [rows, setRows] = useState([]);
  const [rowModesModel, setRowModesModel] = useState({});

  const setRatio = useCallback(
    (ratioRow) => {
      let ratioJson = {};
      ratioRow.forEach((row) => {
        ratioJson[row.model] = [row.input, row.complete];
      });
      onChange({ target: { name: 'ModelRatio', value: JSON.stringify(ratioJson, null, 2) } });
    },
    [onChange]
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
      setRatio(rows.filter((row) => row.id !== id));
    },
    [rows, setRatio]
  );

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

  const processRowUpdate = (newRow, oldRows) => {
    if (!newRow.isNew && newRow.model === oldRows.model && newRow.input === oldRows.input && newRow.complete === oldRows.complete) {
      return oldRows;
    }
    const updatedRow = { ...newRow, isNew: false };
    const error = validation(updatedRow, rows);
    if (error) {
      return Promise.reject(new Error(error));
    }
    setRatio(rows.map((row) => (row.id === newRow.id ? updatedRow : row)));
    return updatedRow;
  };

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
        width: 220,
        editable: true,
        hideable: false
      },
      {
        field: 'input',
        sortable: false,
        headerName: '输入倍率',
        width: 150,
        type: 'number',
        editable: true,
        valueFormatter: (params) => {
          if (params.value == null) {
            return '';
          }
          return `$${parseFloat(params.value * 0.002).toFixed(4)} / ￥${parseFloat(params.value * 0.014).toFixed(4)}`;
        },
        hideable: false
      },
      {
        field: 'complete',
        sortable: false,
        headerName: '完成倍率',
        width: 150,
        type: 'number',
        editable: true,
        valueFormatter: (params) => {
          if (params.value == null) {
            return '';
          }
          return `$${parseFloat(params.value * 0.002).toFixed(4)} / ￥${parseFloat(params.value * 0.014).toFixed(4)}`;
        },
        hideable: false
      },
      {
        field: 'actions',
        type: 'actions',
        headerName: '操作',
        width: 100,
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
    [handleEditClick, handleSaveClick, handleDeleteClick, handleCancelClick, rowModesModel]
  );

  useEffect(() => {
    let modelRatioList = [];
    let itemJson = JSON.parse(ratio);
    let id = 0;
    for (let key in itemJson) {
      modelRatioList.push({ id: id++, model: key, input: itemJson[key][0], complete: itemJson[key][1] });
    }
    setRows(modelRatioList);
  }, [ratio]);

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
        rows={rows}
        columns={modelRatioColumns}
        editMode="row"
        hideFooter
        disableRowSelectionOnClick
        rowModesModel={rowModesModel}
        onRowModesModelChange={handleRowModesModelChange}
        processRowUpdate={processRowUpdate}
        onProcessRowUpdateError={handleProcessRowUpdateError}
        slots={{
          toolbar: EditToolbar
        }}
        slotProps={{
          toolbar: { setRows, setRowModesModel }
        }}
      />
    </Box>
  );
};

ModelRationDataGrid.propTypes = {
  ratio: PropTypes.string.isRequired,
  onChange: PropTypes.func.isRequired
};

export default ModelRationDataGrid;
