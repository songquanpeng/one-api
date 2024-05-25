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

function validation(row) {
  if (row.name === '') {
    return '名称不能为空';
  }

  if (row.url === '') {
    return 'URL不能为空';
  }

  if (row.sort != '' && !/^[0-9]\d*$/.test(row.sort)) {
    return '排序必须为正整数';
  }

  return false;
}

function randomId() {
  return Math.random().toString(36).substr(2, 9);
}

function EditToolbar({ setRows, setRowModesModel }) {
  const handleClick = () => {
    const id = randomId();
    setRows((oldRows) => [{ id, name: '', url: '', show: true, sort: 0, isNew: true }, ...oldRows]);
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

const ChatLinksDataGrid = ({ links, onChange }) => {
  const [rows, setRows] = useState([]);
  const [rowModesModel, setRowModesModel] = useState({});

  const setLinks = useCallback(
    (linksRow) => {
      let linksJson = [];
      // 删除 linksrow 中的 isNew 属性
      linksRow.forEach((row) => {
        let { isNew, ...rest } = row; // eslint-disable-line no-unused-vars
        linksJson.push(rest);
      });
      onChange({ target: { name: 'ChatLinks', value: JSON.stringify(linksJson, null, 2) } });
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
      setLinks(rows.filter((row) => row.id !== id));
    },
    [rows, setLinks]
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
    if (
      !newRow.isNew &&
      newRow.name === oldRows.name &&
      newRow.url === oldRows.url &&
      newRow.sort === oldRows.sort &&
      newRow.show === oldRows.show
    ) {
      return oldRows;
    }
    const updatedRow = { ...newRow, isNew: false };
    const error = validation(updatedRow);
    if (error) {
      return Promise.reject(new Error(error));
    }
    setLinks(rows.map((row) => (row.id === newRow.id ? updatedRow : row)));
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
        field: 'name',
        sortable: true,
        headerName: '名称',
        flex: 1,
        minWidth: 220,
        editable: true,
        hideable: false
      },
      {
        field: 'url',
        sortable: false,
        headerName: '链接',
        flex: 1,
        minWidth: 300,
        editable: true,
        hideable: false
      },
      {
        field: 'show',
        sortable: false,
        headerName: '是否显示在playground',
        flex: 1,
        minWidth: 200,
        type: 'boolean',
        editable: true,
        hideable: false
      },
      {
        field: 'sort',
        sortable: true,
        headerName: '排序',
        type: 'number',
        flex: 1,
        minWidth: 100,
        editable: true,
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
    let itemJson = JSON.parse(links);
    setRows(itemJson);
  }, [links]);

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

ChatLinksDataGrid.propTypes = {
  links: PropTypes.string.isRequired,
  onChange: PropTypes.func.isRequired
};

export default ChatLinksDataGrid;
