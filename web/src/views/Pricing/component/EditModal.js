import PropTypes from 'prop-types';
import * as Yup from 'yup';
import { Formik } from 'formik';
import { useTheme } from '@mui/material/styles';
import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Divider,
  FormControl,
  InputLabel,
  OutlinedInput,
  InputAdornment,
  FormHelperText,
  Select,
  Autocomplete,
  TextField,
  Checkbox,
  MenuItem
} from '@mui/material';

import { showSuccess, showError } from 'utils/common';
import { API } from 'utils/api';
import { createFilterOptions } from '@mui/material/Autocomplete';
import { ValueFormatter, priceType } from './util';
import CheckBoxOutlineBlankIcon from '@mui/icons-material/CheckBoxOutlineBlank';
import CheckBoxIcon from '@mui/icons-material/CheckBox';

const icon = <CheckBoxOutlineBlankIcon fontSize="small" />;
const checkedIcon = <CheckBoxIcon fontSize="small" />;

const filter = createFilterOptions();
const validationSchema = Yup.object().shape({
  is_edit: Yup.boolean(),
  type: Yup.string().oneOf(['tokens', 'times'], '类型 错误').required('类型 不能为空'),
  channel_type: Yup.number().min(1, '渠道类型 错误').required('渠道类型 不能为空'),
  input: Yup.number().required('输入倍率 不能为空'),
  output: Yup.number().required('输出倍率 不能为空'),
  models: Yup.array().min(1, '模型 不能为空')
});

const originInputs = {
  is_edit: false,
  type: 'tokens',
  channel_type: 1,
  input: 0,
  output: 0,
  models: []
};

const EditModal = ({ open, pricesItem, onCancel, onOk, ownedby, noPriceModel }) => {
  const theme = useTheme();
  const [inputs, setInputs] = useState(originInputs);
  const [selectModel, setSelectModel] = useState([]);

  const submit = async (values, { setErrors, setStatus, setSubmitting }) => {
    setSubmitting(true);
    try {
      const res = await API.post(`/api/prices/multiple`, {
        original_models: inputs.models,
        models: values.models,
        price: {
          model: 'batch',
          type: values.type,
          channel_type: values.channel_type,
          input: values.input,
          output: values.output
        }
      });
      const { success, message } = res.data;
      if (success) {
        showSuccess('保存成功！');
        setSubmitting(false);
        setStatus({ success: true });
        onOk(true);
        return;
      } else {
        setStatus({ success: false });
        showError(message);
        setErrors({ submit: message });
      }
    } catch (error) {
      setStatus({ success: false });
      showError(error.message);
      setErrors({ submit: error.message });
      return;
    }
    onOk();
  };

  useEffect(() => {
    if (pricesItem) {
      setSelectModel(pricesItem.models.concat(noPriceModel));
    } else {
      setSelectModel(noPriceModel);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pricesItem, noPriceModel]);

  useEffect(() => {
    if (pricesItem) {
      setInputs(pricesItem);
    } else {
      setInputs(originInputs);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pricesItem]);

  return (
    <Dialog open={open} onClose={onCancel} fullWidth maxWidth={'md'}>
      <DialogTitle sx={{ margin: '0px', fontWeight: 700, lineHeight: '1.55556', padding: '24px', fontSize: '1.125rem' }}>
        {pricesItem ? '编辑' : '新建'}
      </DialogTitle>
      <Divider />
      <DialogContent>
        <Formik initialValues={inputs} enableReinitialize validationSchema={validationSchema} onSubmit={submit}>
          {({ errors, handleBlur, handleChange, handleSubmit, touched, values, isSubmitting }) => (
            <form noValidate onSubmit={handleSubmit}>
              <FormControl fullWidth error={Boolean(touched.type && errors.type)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="type-label">名称</InputLabel>
                <Select
                  id="type-label"
                  label="类型"
                  value={values.type}
                  name="type"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  MenuProps={{
                    PaperProps: {
                      style: {
                        maxHeight: 200
                      }
                    }
                  }}
                >
                  {Object.values(priceType).map((option) => {
                    return (
                      <MenuItem key={option.value} value={option.value}>
                        {option.label}
                      </MenuItem>
                    );
                  })}
                </Select>
                {touched.type && errors.type && (
                  <FormHelperText error id="helper-tex-type-label">
                    {errors.type}
                  </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth error={Boolean(touched.channel_type && errors.channel_type)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="channel_type-label">渠道类型</InputLabel>
                <Select
                  id="channel_type-label"
                  label="渠道类型"
                  value={values.channel_type}
                  name="channel_type"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  MenuProps={{
                    PaperProps: {
                      style: {
                        maxHeight: 200
                      }
                    }
                  }}
                >
                  {Object.values(ownedby).map((option) => {
                    return (
                      <MenuItem key={option.value} value={option.value}>
                        {option.label}
                      </MenuItem>
                    );
                  })}
                </Select>
                {touched.channel_type && errors.channel_type && (
                  <FormHelperText error id="helper-tex-channel_type-label">
                    {errors.channel_type}
                  </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth error={Boolean(touched.input && errors.input)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="channel-input-label">输入倍率</InputLabel>
                <OutlinedInput
                  id="channel-input-label"
                  label="输入倍率"
                  type="number"
                  value={values.input}
                  name="input"
                  endAdornment={<InputAdornment position="end">{ValueFormatter(values.input)}</InputAdornment>}
                  onBlur={handleBlur}
                  onChange={handleChange}
                  aria-describedby="helper-text-channel-input-label"
                />

                {touched.input && errors.input && (
                  <FormHelperText error id="helper-tex-channel-input-label">
                    {errors.input}
                  </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth error={Boolean(touched.output && errors.output)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="channel-output-label">输出倍率</InputLabel>
                <OutlinedInput
                  id="channel-output-label"
                  label="输出倍率"
                  type="number"
                  value={values.output}
                  name="output"
                  endAdornment={<InputAdornment position="end">{ValueFormatter(values.output)}</InputAdornment>}
                  onBlur={handleBlur}
                  onChange={handleChange}
                  aria-describedby="helper-text-channel-output-label"
                />

                {touched.output && errors.output && (
                  <FormHelperText error id="helper-tex-channel-output-label">
                    {errors.output}
                  </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth sx={{ ...theme.typography.otherInput }}>
                <Autocomplete
                  multiple
                  freeSolo
                  id="channel-models-label"
                  options={selectModel}
                  value={values.models}
                  onChange={(e, value) => {
                    const event = {
                      target: {
                        name: 'models',
                        value: value
                      }
                    };
                    handleChange(event);
                  }}
                  onBlur={handleBlur}
                  // filterSelectedOptions
                  disableCloseOnSelect
                  renderInput={(params) => <TextField {...params} name="models" error={Boolean(errors.models)} label="模型" />}
                  filterOptions={(options, params) => {
                    const filtered = filter(options, params);
                    const { inputValue } = params;
                    const isExisting = options.some((option) => inputValue === option);
                    if (inputValue !== '' && !isExisting) {
                      filtered.push(inputValue);
                    }
                    return filtered;
                  }}
                  renderOption={(props, option, { selected }) => (
                    <li {...props}>
                      <Checkbox icon={icon} checkedIcon={checkedIcon} style={{ marginRight: 8 }} checked={selected} />
                      {option}
                    </li>
                  )}
                />
                {errors.models ? (
                  <FormHelperText error id="helper-tex-channel-models-label">
                    {errors.models}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-models-label">
                    {' '}
                    请选择该价格所支持的模型,你也可以输入通配符*来匹配模型，例如：gpt-3.5*，表示支持所有gpt-3.5开头的模型，*号只能在最后一位使用，前面必须有字符，例如：gpt-3.5*是正确的，*gpt-3.5是错误的{' '}
                  </FormHelperText>
                )}
              </FormControl>

              <DialogActions>
                <Button onClick={onCancel}>取消</Button>
                <Button disableElevation disabled={isSubmitting} type="submit" variant="contained" color="primary">
                  提交
                </Button>
              </DialogActions>
            </form>
          )}
        </Formik>
      </DialogContent>
    </Dialog>
  );
};

export default EditModal;

EditModal.propTypes = {
  open: PropTypes.bool,
  pricesItem: PropTypes.object,
  onCancel: PropTypes.func,
  onOk: PropTypes.func,
  ownedby: PropTypes.array,
  noPriceModel: PropTypes.array
};
