import { useState } from 'react';
import { Grid, TextField, InputAdornment, Checkbox, Button, FormControlLabel, IconButton, Alert } from '@mui/material';
import { gridSpacing } from 'store/constant';
import { IconSearch, IconHttpDelete } from '@tabler/icons-react';
import { fetchChannelData } from '../index';
import { API } from 'utils/api';
import { showError, showSuccess } from 'utils/common';

const BatchDelModel = () => {
  const [value, setValue] = useState('');
  const [data, setData] = useState([]);
  const [selected, setSelected] = useState([]);
  const [loadding, setLoadding] = useState(false);

  const handleSearch = async () => {
    const data = await fetchChannelData(0, 100, { models: value }, 'desc', 'id');
    if (data) {
      // 遍历data 逗号分隔models， 检测是否只有一个model 如果是则排除
      const newData = data.data.filter((item) => {
        if (item.models.split(',').length > 1) {
          return true;
        }
        return false;
      });

      setData(newData);
    }
  };

  const handleSelect = (id) => {
    setSelected((prev) => {
      if (prev.includes(id)) {
        return prev.filter((i) => i !== id);
      } else {
        return [...prev, id];
      }
    });
  };

  const handleSelectAll = () => {
    if (selected.length === data.length) {
      setSelected([]);
    } else {
      setSelected(data.map((item) => item.id));
    }
  };

  const handleSubmit = async () => {
    if (value === '' || selected.length === 0) {
      return;
    }
    setLoadding(true);
    try {
      const res = await API.put(`/api/channel/batch/del_model`, {
        ids: selected,
        value: value
      });

      const { success, message, data } = res.data;
      if (success) {
        showSuccess('成功删除' + data + '条数据');
      } else {
        showError(message);
      }
    } catch (error) {
      console.error(error);
    }
    setLoadding(false);
  };

  return (
    <Grid container spacing={gridSpacing}>
      <Grid item xs={12}>
        <Alert severity="info">如果渠道只有一个模型的，将不会显示，请手动去列表删除渠道</Alert>
      </Grid>
      <Grid item xs={12}>
        <TextField
          sx={{ ml: 1, flex: 1 }}
          placeholder="请输入完整模型名称"
          inputProps={{ 'aria-label': '请输入完整模型名称' }}
          value={value}
          onChange={(e) => {
            setValue(e.target.value);
          }}
          InputProps={{
            endAdornment: (
              <InputAdornment position="end">
                <IconButton aria-label="toggle password visibility" onClick={handleSearch} edge="end">
                  <IconSearch />
                </IconButton>
              </InputAdornment>
            )
          }}
        />
      </Grid>
      {data.length === 0 ? (
        <Grid item xs={12}>
          暂无数据
        </Grid>
      ) : (
        <>
          <Grid item xs={12}>
            <Button onClick={handleSelectAll}>{selected.length === data.length ? '反全选' : '全选'}</Button>
          </Grid>
          <Grid item xs={12}>
            {data.map((item) => (
              <FormControlLabel
                key={item.id}
                control={<Checkbox checked={selected.includes(item.id)} onChange={() => handleSelect(item.id)} />}
                label={item.name}
              />
            ))}
          </Grid>
          <Grid item xs={12}>
            <Button variant="contained" color="primary" startIcon={<IconHttpDelete />} onClick={handleSubmit} disabled={loadding}>
              删除
            </Button>
          </Grid>
        </>
      )}
    </Grid>
  );
};

export default BatchDelModel;
