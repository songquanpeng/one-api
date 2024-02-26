import PropTypes from 'prop-types';
import { useTheme } from '@mui/material/styles';
import { IconKey, IconBrandGithubCopilot, IconSitemap, IconVersions } from '@tabler/icons-react';
import { InputAdornment, OutlinedInput, Stack, FormControl, InputLabel, Select, MenuItem } from '@mui/material'; //
import { CHANNEL_OPTIONS } from 'constants/ChannelConstants';
// ----------------------------------------------------------------------

export default function TableToolBar({ filterName, handleFilterName, groupOptions }) {
  const theme = useTheme();
  const grey500 = theme.palette.grey[500];

  return (
    <>
      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={{ xs: 3, sm: 2, md: 4 }} padding={'24px'} paddingBottom={'0px'}>
        <FormControl>
          <InputLabel htmlFor="channel-name-label">渠道名称</InputLabel>
          <OutlinedInput
            id="name"
            name="name"
            sx={{
              minWidth: '100%'
            }}
            label="渠道名称"
            value={filterName.name}
            onChange={handleFilterName}
            placeholder="渠道名称"
            startAdornment={
              <InputAdornment position="start">
                <IconSitemap stroke={1.5} size="20px" color={grey500} />
              </InputAdornment>
            }
          />
        </FormControl>
        <FormControl>
          <InputLabel htmlFor="channel-models-label">模型名称</InputLabel>
          <OutlinedInput
            id="models"
            name="models"
            sx={{
              minWidth: '100%'
            }}
            label="模型名称"
            value={filterName.models}
            onChange={handleFilterName}
            placeholder="模型名称"
            startAdornment={
              <InputAdornment position="start">
                <IconBrandGithubCopilot stroke={1.5} size="20px" color={grey500} />
              </InputAdornment>
            }
          />
        </FormControl>
        <FormControl>
          <InputLabel htmlFor="channel-test_model-label">测试模型</InputLabel>
          <OutlinedInput
            id="test_model"
            name="test_model"
            sx={{
              minWidth: '100%'
            }}
            label="测试模型"
            value={filterName.test_model}
            onChange={handleFilterName}
            placeholder="测试模型"
            startAdornment={
              <InputAdornment position="start">
                <IconBrandGithubCopilot stroke={1.5} size="20px" color={grey500} />
              </InputAdornment>
            }
          />
        </FormControl>
        <FormControl>
          <InputLabel htmlFor="channel-key-label">key</InputLabel>
          <OutlinedInput
            id="key"
            name="key"
            sx={{
              minWidth: '100%'
            }}
            label="key"
            value={filterName.key}
            onChange={handleFilterName}
            placeholder="key"
            startAdornment={
              <InputAdornment position="start">
                <IconKey stroke={1.5} size="20px" color={grey500} />
              </InputAdornment>
            }
          />
        </FormControl>
        <FormControl>
          <InputLabel htmlFor="channel-other-label">其他参数</InputLabel>
          <OutlinedInput
            id="other"
            name="other"
            sx={{
              minWidth: '100%'
            }}
            label="其他参数"
            value={filterName.other}
            onChange={handleFilterName}
            placeholder="其他参数"
            startAdornment={
              <InputAdornment position="start">
                <IconVersions stroke={1.5} size="20px" color={grey500} />
              </InputAdornment>
            }
          />
        </FormControl>
      </Stack>

      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={{ xs: 3, sm: 2, md: 4 }} padding={'24px'}>
        <FormControl sx={{ minWidth: '22%' }}>
          <InputLabel htmlFor="channel-type-label">渠道类型</InputLabel>
          <Select
            id="channel-type-label"
            label="渠道类型"
            value={filterName.type}
            name="type"
            onChange={handleFilterName}
            sx={{
              minWidth: '100%'
            }}
            MenuProps={{
              PaperProps: {
                style: {
                  maxHeight: 200
                }
              }
            }}
          >
            <MenuItem key={0} value={0}>
              全部
            </MenuItem>

            {Object.values(CHANNEL_OPTIONS).map((option) => {
              return (
                <MenuItem key={option.value} value={option.value}>
                  {option.text}
                </MenuItem>
              );
            })}
          </Select>
        </FormControl>
        <FormControl sx={{ minWidth: '22%' }}>
          <InputLabel htmlFor="channel-status-label">状态</InputLabel>
          <Select
            id="channel-status-label"
            label="状态"
            value={filterName.status}
            name="status"
            onChange={handleFilterName}
            sx={{
              minWidth: '100%'
            }}
            MenuProps={{
              PaperProps: {
                style: {
                  maxHeight: 200
                }
              }
            }}
          >
            <MenuItem key={0} value={0}>
              全部
            </MenuItem>
            <MenuItem key={1} value={1}>
              启用
            </MenuItem>
            <MenuItem key={2} value={2}>
              禁用
            </MenuItem>
            <MenuItem key={3} value={3}>
              测速禁用
            </MenuItem>
          </Select>
        </FormControl>

        <FormControl sx={{ minWidth: '22%' }}>
          <InputLabel htmlFor="channel-group-label">分组</InputLabel>
          <Select
            id="channel-group-label"
            label="分组"
            value={filterName.group}
            name="group"
            onChange={handleFilterName}
            sx={{
              minWidth: '100%'
            }}
            MenuProps={{
              PaperProps: {
                style: {
                  maxHeight: 200
                }
              }
            }}
          >
            {groupOptions.map((option) => {
              return (
                <MenuItem key={option} value={option}>
                  {option}
                </MenuItem>
              );
            })}
          </Select>
        </FormControl>
      </Stack>
    </>
  );
}

TableToolBar.propTypes = {
  filterName: PropTypes.object,
  handleFilterName: PropTypes.func,
  groupOptions: PropTypes.array
};
