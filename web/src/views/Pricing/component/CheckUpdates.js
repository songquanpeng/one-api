import PropTypes from 'prop-types';
import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Divider,
  Button,
  TextField,
  Grid,
  FormControl,
  Alert,
  Stack,
  Typography
} from '@mui/material';
import { API } from 'utils/api';
import { showError, showSuccess } from 'utils/common';
import LoadingButton from '@mui/lab/LoadingButton';
import Label from 'ui-component/Label';

export const CheckUpdates = ({ open, onCancel, onOk, row }) => {
  const [url, setUrl] = useState('https://raw.githubusercontent.com/MartialBE/one-api/prices/prices.json');
  const [loading, setLoading] = useState(false);
  const [updateLoading, setUpdateLoading] = useState(false);
  const [newPricing, setNewPricing] = useState([]);
  const [addModel, setAddModel] = useState([]);
  const [diffModel, setDiffModel] = useState([]);

  const handleCheckUpdates = async () => {
    setLoading(true);
    try {
      const res = await API.get(url);
      // 检测是否是一个列表
      if (!Array.isArray(res.data)) {
        showError('数据格式不正确');
      } else {
        setNewPricing(res.data);
      }
    } catch (err) {
      console.error(err);
    }
    setLoading(false);
  };

  const syncPricing = async (overwrite) => {
    setUpdateLoading(true);
    if (!newPricing.length) {
      showError('请先获取数据');
      return;
    }

    if (!overwrite && !addModel.length) {
      showError('没有新增模型');
      return;
    }
    try {
      overwrite = overwrite ? 'true' : 'false';
      const res = await API.post('/api/prices/sync?overwrite=' + overwrite, newPricing);
      const { success, message } = res.data;
      if (success) {
        showSuccess('操作成功完成！');
        onOk(true);
      } else {
        showError(message);
      }
    } catch (err) {
      console.error(err);
    }
    setUpdateLoading(false);
  };

  useEffect(() => {
    const newModels = newPricing.filter((np) => !row.some((r) => r.model === np.model));

    const changeModel = row.filter((r) =>
      newPricing.some((np) => np.model === r.model && (np.input !== r.input || np.output !== r.output))
    );

    if (newModels.length > 0) {
      const newModelsList = newModels.map((model) => model.model);
      setAddModel(newModelsList);
    } else {
      setAddModel('');
    }

    if (changeModel.length > 0) {
      const changeModelList = changeModel.map((model) => {
        const newModel = newPricing.find((np) => np.model === model.model);
        let changes = '';
        if (model.input !== newModel.input) {
          changes += `输入倍率由 ${model.input} 变为 ${newModel.input},`;
        }
        if (model.output !== newModel.output) {
          changes += `输出倍率由 ${model.output} 变为 ${newModel.output}`;
        }
        return `${model.model}:${changes}`;
      });
      setDiffModel(changeModelList);
    } else {
      setDiffModel('');
    }
  }, [row, newPricing]);

  return (
    <Dialog open={open} onClose={onCancel} fullWidth maxWidth={'md'}>
      <DialogTitle sx={{ margin: '0px', fontWeight: 700, lineHeight: '1.55556', padding: '24px', fontSize: '1.125rem' }}>
        检查更新
      </DialogTitle>
      <Divider />
      <DialogContent>
        <Grid container justifyContent="center" alignItems="center" spacing={2}>
          <Grid item xs={12} md={10}>
            <FormControl fullWidth component="fieldset">
              <TextField label="URL" variant="outlined" value={url} onChange={(e) => setUrl(e.target.value)} />
            </FormControl>
          </Grid>
          <Grid item xs={12} md={2}>
            <LoadingButton variant="contained" color="primary" onClick={handleCheckUpdates} loading={loading}>
              获取数据
            </LoadingButton>
          </Grid>
          {newPricing.length > 0 && (
            <Grid item xs={12}>
              {!addModel.length && !diffModel.length && <Alert severity="success">无更新</Alert>}

              {addModel.length > 0 && (
                <Alert severity="warning">
                  新增模型：
                  <Stack direction="row" spacing={1} flexWrap="wrap">
                    {addModel.map((model) => (
                      <Label color="info" key={model} variant="outlined">
                        {model}
                      </Label>
                    ))}
                  </Stack>
                </Alert>
              )}

              {diffModel.length > 0 && (
                <Alert severity="warning">
                  价格变动模型(仅供参考，如果你自己修改了对应模型的价格请忽略)：
                  {diffModel.map((model) => (
                    <Typography variant="button" display="block" gutterBottom key={model}>
                      {model}
                    </Typography>
                  ))}
                </Alert>
              )}
              <Alert severity="warning">
                注意:
                你可以选择覆盖或者仅添加新增，如果你选择覆盖，将会删除你自己添加的模型价格，完全使用远程配置，如果你选择仅添加新增，将会只会添加
                新增模型的价格
              </Alert>
              <Stack direction="row" justifyContent="center" spacing={1} flexWrap="wrap">
                <LoadingButton
                  variant="contained"
                  color="primary"
                  onClick={() => {
                    syncPricing(true);
                  }}
                  loading={updateLoading}
                >
                  覆盖数据
                </LoadingButton>
                <LoadingButton
                  variant="contained"
                  color="primary"
                  onClick={() => {
                    syncPricing(false);
                  }}
                  loading={updateLoading}
                >
                  仅添加新增
                </LoadingButton>
              </Stack>
            </Grid>
          )}
        </Grid>
      </DialogContent>
      <DialogActions>
        <Button onClick={onCancel} color="primary">
          取消
        </Button>
      </DialogActions>
    </Dialog>
  );
};

CheckUpdates.propTypes = {
  open: PropTypes.bool,
  row: PropTypes.array,
  onCancel: PropTypes.func,
  onOk: PropTypes.func
};
