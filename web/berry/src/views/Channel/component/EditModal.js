import PropTypes from "prop-types";
import { useState, useEffect } from "react";
import { CHANNEL_OPTIONS } from "constants/ChannelConstants";
import { useTheme } from "@mui/material/styles";
import { API } from "utils/api";
import { showError, showSuccess } from "utils/common";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Button,
  Divider,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  OutlinedInput,
  ButtonGroup,
  Container,
  Autocomplete,
  FormHelperText,
} from "@mui/material";

import { Formik } from "formik";
import * as Yup from "yup";
import { defaultConfig, typeConfig } from "../type/Config"; //typeConfig
import { createFilterOptions } from "@mui/material/Autocomplete";

const filter = createFilterOptions();
const validationSchema = Yup.object().shape({
  is_edit: Yup.boolean(),
  name: Yup.string().required("名称 不能为空"),
  type: Yup.number().required("渠道 不能为空"),
  key: Yup.string().when("is_edit", {
    is: false,
    then: Yup.string().required("密钥 不能为空"),
  }),
  other: Yup.string(),
  proxy: Yup.string(),
  test_model: Yup.string(),
  models: Yup.array().min(1, "模型 不能为空"),
  groups: Yup.array().min(1, "用户组 不能为空"),
  base_url: Yup.string().when("type", {
    is: (value) => [3, 24, 8].includes(value),
    then: Yup.string().required("渠道API地址 不能为空"), // base_url 是必需的
    otherwise: Yup.string(), // 在其他情况下，base_url 可以是任意字符串
  }),
  model_mapping: Yup.string().test(
    "is-json",
    "必须是有效的JSON字符串",
    function (value) {
      try {
        if (value === "" || value === null || value === undefined) {
          return true;
        }
        const parsedValue = JSON.parse(value);
        if (typeof parsedValue === "object") {
          return true;
        }
      } catch (e) {
        return false;
      }
      return false;
    }
  ),
});

const EditModal = ({ open, channelId, onCancel, onOk }) => {
  const theme = useTheme();
  // const [loading, setLoading] = useState(false);
  const [initialInput, setInitialInput] = useState(defaultConfig.input);
  const [inputLabel, setInputLabel] = useState(defaultConfig.inputLabel); //
  const [inputPrompt, setInputPrompt] = useState(defaultConfig.prompt);
  const [groupOptions, setGroupOptions] = useState([]);
  const [modelOptions, setModelOptions] = useState([]);

  const initChannel = (typeValue) => {
    if (typeConfig[typeValue]?.inputLabel) {
      setInputLabel({
        ...defaultConfig.inputLabel,
        ...typeConfig[typeValue].inputLabel,
      });
    } else {
      setInputLabel(defaultConfig.inputLabel);
    }

    if (typeConfig[typeValue]?.prompt) {
      setInputPrompt({
        ...defaultConfig.prompt,
        ...typeConfig[typeValue].prompt,
      });
    } else {
      setInputPrompt(defaultConfig.prompt);
    }

    return typeConfig[typeValue]?.input;
  };
  const handleTypeChange = (setFieldValue, typeValue, values) => {
    const newInput = initChannel(typeValue);

    if (newInput) {
      Object.keys(newInput).forEach((key) => {
        if (
          (!Array.isArray(values[key]) &&
            values[key] !== null &&
            values[key] !== undefined &&
            values[key] !== "") ||
          (Array.isArray(values[key]) && values[key].length > 0)
        ) {
          return;
        }

        if (key === "models") {
          setFieldValue(key, initialModel(newInput[key]));
          return;
        }
        setFieldValue(key, newInput[key]);
      });
    }
  };

  const basicModels = (channelType) => {
    let modelGroup =
      typeConfig[channelType]?.modelGroup || defaultConfig.modelGroup;
    // 循环 modelOptions，找到 modelGroup 对应的模型
    let modelList = [];
    modelOptions.forEach((model) => {
      if (model.group === modelGroup) {
        modelList.push(model);
      }
    });
    return modelList;
  };

  const fetchGroups = async () => {
    try {
      let res = await API.get(`/api/group/`);
      setGroupOptions(res.data.data);
    } catch (error) {
      showError(error.message);
    }
  };

  const fetchModels = async () => {
    try {
      let res = await API.get(`/api/channel/models`);
      setModelOptions(
        res.data.data.map((model) => {
          return {
            id: model.id,
            group: model.owned_by,
          };
        })
      );
    } catch (error) {
      showError(error.message);
    }
  };

  const submit = async (values, { setErrors, setStatus, setSubmitting }) => {
    setSubmitting(true);
    if (values.base_url && values.base_url.endsWith("/")) {
      values.base_url = values.base_url.slice(0, values.base_url.length - 1);
    }
    if (values.type === 3 && values.other === "") {
      values.other = "2023-09-01-preview";
    }
    if (values.type === 18 && values.other === "") {
      values.other = "v2.1";
    }
    let res;
    const modelsStr = values.models.map((model) => model.id).join(",");
    values.group = values.groups.join(",");
    if (channelId) {
      res = await API.put(`/api/channel/`, {
        ...values,
        id: parseInt(channelId),
        models: modelsStr,
      });
    } else {
      res = await API.post(`/api/channel/`, { ...values, models: modelsStr });
    }
    const { success, message } = res.data;
    if (success) {
      if (channelId) {
        showSuccess("渠道更新成功！");
      } else {
        showSuccess("渠道创建成功！");
      }
      setSubmitting(false);
      setStatus({ success: true });
      onOk(true);
    } else {
      setStatus({ success: false });
      showError(message);
      setErrors({ submit: message });
    }
  };

  function initialModel(channelModel) {
    if (!channelModel) {
      return [];
    }

    // 如果 channelModel 是一个字符串
    if (typeof channelModel === "string") {
      channelModel = channelModel.split(",");
    }
    let modelList = channelModel.map((model) => {
      const modelOption = modelOptions.find((option) => option.id === model);
      if (modelOption) {
        return modelOption;
      }
      return { id: model, group: "自定义：点击或回车输入" };
    });
    return modelList;
  }

  const loadChannel = async () => {
    let res = await API.get(`/api/channel/${channelId}`);
    const { success, message, data } = res.data;
    if (success) {
      if (data.models === "") {
        data.models = [];
      } else {
        data.models = initialModel(data.models);
      }
      if (data.group === "") {
        data.groups = [];
      } else {
        data.groups = data.group.split(",");
      }
      if (data.model_mapping !== "") {
        data.model_mapping = JSON.stringify(
          JSON.parse(data.model_mapping),
          null,
          2
        );
      }
      data.is_edit = true;
      initChannel(data.type);
      setInitialInput(data);
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    fetchGroups().then();
    fetchModels().then();
    if (channelId) {
      loadChannel().then();
    } else {
      initChannel(1);
      setInitialInput({ ...defaultConfig.input, is_edit: false });
    }
  }, [channelId]);

  return (
    <Dialog open={open} onClose={onCancel} fullWidth maxWidth={"md"}>
      <DialogTitle
        sx={{
          margin: "0px",
          fontWeight: 700,
          lineHeight: "1.55556",
          padding: "24px",
          fontSize: "1.125rem",
        }}
      >
        {channelId ? "编辑渠道" : "新建渠道"}
      </DialogTitle>
      <Divider />
      <DialogContent>
        <Formik
          initialValues={initialInput}
          enableReinitialize
          validationSchema={validationSchema}
          onSubmit={submit}
        >
          {({
            errors,
            handleBlur,
            handleChange,
            handleSubmit,
            isSubmitting,
            touched,
            values,
            setFieldValue,
          }) => (
            <form noValidate onSubmit={handleSubmit}>
              <FormControl
                fullWidth
                error={Boolean(touched.type && errors.type)}
                sx={{ ...theme.typography.otherInput }}
              >
                <InputLabel htmlFor="channel-type-label">
                  {inputLabel.type}
                </InputLabel>
                <Select
                  id="channel-type-label"
                  label={inputLabel.type}
                  value={values.type}
                  name="type"
                  onBlur={handleBlur}
                  onChange={(e) => {
                    handleChange(e);
                    handleTypeChange(setFieldValue, e.target.value, values);
                  }}
                  MenuProps={{
                    PaperProps: {
                      style: {
                        maxHeight: 200,
                      },
                    },
                  }}
                >
                  {Object.values(CHANNEL_OPTIONS).map((option) => {
                    return (
                      <MenuItem key={option.value} value={option.value}>
                        {option.text}
                      </MenuItem>
                    );
                  })}
                </Select>
                {touched.type && errors.type ? (
                  <FormHelperText error id="helper-tex-channel-type-label">
                    {errors.type}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-type-label">
                    {" "}
                    {inputPrompt.type}{" "}
                  </FormHelperText>
                )}
              </FormControl>

              <FormControl
                fullWidth
                error={Boolean(touched.name && errors.name)}
                sx={{ ...theme.typography.otherInput }}
              >
                <InputLabel htmlFor="channel-name-label">
                  {inputLabel.name}
                </InputLabel>
                <OutlinedInput
                  id="channel-name-label"
                  label={inputLabel.name}
                  type="text"
                  value={values.name}
                  name="name"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{ autoComplete: "name" }}
                  aria-describedby="helper-text-channel-name-label"
                />
                {touched.name && errors.name ? (
                  <FormHelperText error id="helper-tex-channel-name-label">
                    {errors.name}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-name-label">
                    {" "}
                    {inputPrompt.name}{" "}
                  </FormHelperText>
                )}
              </FormControl>

              <FormControl
                fullWidth
                error={Boolean(touched.base_url && errors.base_url)}
                sx={{ ...theme.typography.otherInput }}
              >
                <InputLabel htmlFor="channel-base_url-label">
                  {inputLabel.base_url}
                </InputLabel>
                <OutlinedInput
                  id="channel-base_url-label"
                  label={inputLabel.base_url}
                  type="text"
                  value={values.base_url}
                  name="base_url"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{}}
                  aria-describedby="helper-text-channel-base_url-label"
                />
                {touched.base_url && errors.base_url ? (
                  <FormHelperText error id="helper-tex-channel-base_url-label">
                    {errors.base_url}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-base_url-label">
                    {" "}
                    {inputPrompt.base_url}{" "}
                  </FormHelperText>
                )}
              </FormControl>

              {inputPrompt.other && (
                <FormControl
                  fullWidth
                  error={Boolean(touched.other && errors.other)}
                  sx={{ ...theme.typography.otherInput }}
                >
                  <InputLabel htmlFor="channel-other-label">
                    {inputLabel.other}
                  </InputLabel>
                  <OutlinedInput
                    id="channel-other-label"
                    label={inputLabel.other}
                    type="text"
                    value={values.other}
                    name="other"
                    onBlur={handleBlur}
                    onChange={handleChange}
                    inputProps={{}}
                    aria-describedby="helper-text-channel-other-label"
                  />
                  {touched.other && errors.other ? (
                    <FormHelperText error id="helper-tex-channel-other-label">
                      {errors.other}
                    </FormHelperText>
                  ) : (
                    <FormHelperText id="helper-tex-channel-other-label">
                      {" "}
                      {inputPrompt.other}{" "}
                    </FormHelperText>
                  )}
                </FormControl>
              )}

              <FormControl fullWidth sx={{ ...theme.typography.otherInput }}>
                <Autocomplete
                  multiple
                  id="channel-groups-label"
                  options={groupOptions}
                  value={values.groups}
                  onChange={(e, value) => {
                    const event = {
                      target: {
                        name: "groups",
                        value: value,
                      },
                    };
                    handleChange(event);
                  }}
                  onBlur={handleBlur}
                  filterSelectedOptions
                  renderInput={(params) => (
                    <TextField
                      {...params}
                      name="groups"
                      error={Boolean(errors.groups)}
                      label={inputLabel.groups}
                    />
                  )}
                  aria-describedby="helper-text-channel-groups-label"
                />
                {errors.groups ? (
                  <FormHelperText error id="helper-tex-channel-groups-label">
                    {errors.groups}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-groups-label">
                    {" "}
                    {inputPrompt.groups}{" "}
                  </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth sx={{ ...theme.typography.otherInput }}>
                <Autocomplete
                  multiple
                  freeSolo
                  id="channel-models-label"
                  options={modelOptions}
                  value={values.models}
                  onChange={(e, value) => {
                    const event = {
                      target: {
                        name: "models",
                        value: value.map((item) =>
                          typeof item === "string"
                            ? { id: item, group: "自定义：点击或回车输入" }
                            : item
                        ),
                      },
                    };
                    handleChange(event);
                  }}
                  onBlur={handleBlur}
                  filterSelectedOptions
                  renderInput={(params) => (
                    <TextField
                      {...params}
                      name="models"
                      error={Boolean(errors.models)}
                      label={inputLabel.models}
                    />
                  )}
                  groupBy={(option) => option.group}
                  getOptionLabel={(option) => {
                    if (typeof option === "string") {
                      return option;
                    }
                    if (option.inputValue) {
                      return option.inputValue;
                    }
                    return option.id;
                  }}
                  filterOptions={(options, params) => {
                    const filtered = filter(options, params);
                    const { inputValue } = params;
                    const isExisting = options.some(
                      (option) => inputValue === option.id
                    );
                    if (inputValue !== "" && !isExisting) {
                      filtered.push({
                        id: inputValue,
                        group: "自定义：点击或回车输入",
                      });
                    }
                    return filtered;
                  }}
                />
                {errors.models ? (
                  <FormHelperText error id="helper-tex-channel-models-label">
                    {errors.models}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-models-label">
                    {" "}
                    {inputPrompt.models}{" "}
                  </FormHelperText>
                )}
              </FormControl>
              <Container
                sx={{
                  textAlign: "right",
                }}
              >
                <ButtonGroup
                  variant="outlined"
                  aria-label="small outlined primary button group"
                >
                  <Button
                    onClick={() => {
                      setFieldValue("models", basicModels(values.type));
                    }}
                  >
                    填入渠道支持模型
                  </Button>
                  <Button
                    onClick={() => {
                      setFieldValue("models", modelOptions);
                    }}
                  >
                    填入所有模型
                  </Button>
                </ButtonGroup>
              </Container>
              <FormControl
                fullWidth
                error={Boolean(touched.key && errors.key)}
                sx={{ ...theme.typography.otherInput }}
              >
                <InputLabel htmlFor="channel-key-label">
                  {inputLabel.key}
                </InputLabel>
                <OutlinedInput
                  id="channel-key-label"
                  label={inputLabel.key}
                  type="text"
                  value={values.key}
                  name="key"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{}}
                  aria-describedby="helper-text-channel-key-label"
                />
                {touched.key && errors.key ? (
                  <FormHelperText error id="helper-tex-channel-key-label">
                    {errors.key}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-key-label">
                    {" "}
                    {inputPrompt.key}{" "}
                  </FormHelperText>
                )}
              </FormControl>
              <FormControl
                fullWidth
                error={Boolean(touched.model_mapping && errors.model_mapping)}
                sx={{ ...theme.typography.otherInput }}
              >
                {/* <InputLabel htmlFor="channel-model_mapping-label">{inputLabel.model_mapping}</InputLabel> */}
                <TextField
                  multiline
                  id="channel-model_mapping-label"
                  label={inputLabel.model_mapping}
                  value={values.model_mapping}
                  name="model_mapping"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  aria-describedby="helper-text-channel-model_mapping-label"
                  minRows={5}
                  placeholder={inputPrompt.model_mapping}
                />
                {touched.model_mapping && errors.model_mapping ? (
                  <FormHelperText
                    error
                    id="helper-tex-channel-model_mapping-label"
                  >
                    {errors.model_mapping}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-model_mapping-label">
                    {" "}
                    {inputPrompt.model_mapping}{" "}
                  </FormHelperText>
                )}
              </FormControl>
              <FormControl
                fullWidth
                error={Boolean(touched.proxy && errors.proxy)}
                sx={{ ...theme.typography.otherInput }}
              >
                <InputLabel htmlFor="channel-proxy-label">
                  {inputLabel.proxy}
                </InputLabel>
                <OutlinedInput
                  id="channel-proxy-label"
                  label={inputLabel.proxy}
                  type="text"
                  value={values.proxy}
                  name="proxy"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{}}
                  aria-describedby="helper-text-channel-proxy-label"
                />
                {touched.proxy && errors.proxy ? (
                  <FormHelperText error id="helper-tex-channel-proxy-label">
                    {errors.proxy}
                  </FormHelperText>
                ) : (
                  <FormHelperText id="helper-tex-channel-proxy-label">
                    {" "}
                    {inputPrompt.proxy}{" "}
                  </FormHelperText>
                )}
              </FormControl>
              {inputPrompt.test_model && (
                <FormControl
                  fullWidth
                  error={Boolean(touched.test_model && errors.test_model)}
                  sx={{ ...theme.typography.otherInput }}
                >
                  <InputLabel htmlFor="channel-test_model-label">
                    {inputLabel.test_model}
                  </InputLabel>
                  <OutlinedInput
                    id="channel-test_model-label"
                    label={inputLabel.test_model}
                    type="text"
                    value={values.test_model}
                    name="test_model"
                    onBlur={handleBlur}
                    onChange={handleChange}
                    inputProps={{}}
                    aria-describedby="helper-text-channel-test_model-label"
                  />
                  {touched.test_model && errors.test_model ? (
                    <FormHelperText
                      error
                      id="helper-tex-channel-test_model-label"
                    >
                      {errors.test_model}
                    </FormHelperText>
                  ) : (
                    <FormHelperText id="helper-tex-channel-test_model-label">
                      {" "}
                      {inputPrompt.test_model}{" "}
                    </FormHelperText>
                  )}
                </FormControl>
              )}
              <DialogActions>
                <Button onClick={onCancel}>取消</Button>
                <Button
                  disableElevation
                  disabled={isSubmitting}
                  type="submit"
                  variant="contained"
                  color="primary"
                >
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
  channelId: PropTypes.number,
  onCancel: PropTypes.func,
  onOk: PropTypes.func,
};
