import { useState, useEffect } from 'react';
import SubCard from 'ui-component/cards/SubCard';
import {
    Stack,
    FormControl,
    InputLabel,
    OutlinedInput,
    Button,
    Alert,
    TextField,
    Dialog,
    DialogTitle,
    DialogActions,
    DialogContent,
    Divider, Link
} from '@mui/material';
import Grid from '@mui/material/Unstable_Grid2';
import { showError, showSuccess } from 'utils/common'; //,
import { API } from 'utils/api';
import { marked } from 'marked';

const OtherSetting = () => {
  let [inputs, setInputs] = useState({
    Footer: '',
    Notice: '',
    About: '',
    SystemName: '',
    Logo: '',
    HomePageContent: '',
    Theme: '',
  });
  let [loading, setLoading] = useState(false);
  const [showUpdateModal, setShowUpdateModal] = useState(false);
  const [updateData, setUpdateData] = useState({
    tag_name: '',
    content: ''
  });

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        if (item.key in inputs) {
          newInputs[item.key] = item.value;
        }
      });
      setInputs(newInputs);
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    getOptions().then();
  }, []);

  const updateOption = async (key, value) => {
    setLoading(true);
    const res = await API.put('/api/option/', {
      key,
      value
    });
    const { success, message } = res.data;
    if (success) {
      setInputs((inputs) => ({ ...inputs, [key]: value }));
      showSuccess('保存成功');
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const handleInputChange = async (event) => {
    let { name, value } = event.target;
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const submitNotice = async () => {
    await updateOption('Notice', inputs.Notice);
  };

  const submitFooter = async () => {
    await updateOption('Footer', inputs.Footer);
  };

  const submitSystemName = async () => {
    await updateOption('SystemName', inputs.SystemName);
  };

  const submitTheme = async () => {
    await updateOption('Theme', inputs.Theme);
  };

  const submitLogo = async () => {
    await updateOption('Logo', inputs.Logo);
  };

  const submitAbout = async () => {
    await updateOption('About', inputs.About);
  };

  const submitOption = async (key) => {
    await updateOption(key, inputs[key]);
  };

  const openGitHubRelease = () => {
    window.location = 'https://github.com/songquanpeng/one-api/releases/latest';
  };

  const checkUpdate = async () => {
    const res = await API.get('https://api.github.com/repos/songquanpeng/one-api/releases/latest');
    const { tag_name, body } = res.data;
    if (tag_name === process.env.REACT_APP_VERSION) {
      showSuccess(`已是最新版本：${tag_name}`);
    } else {
      setUpdateData({
        tag_name: tag_name,
        content: marked.parse(body)
      });
      setShowUpdateModal(true);
    }
  };

  return (
    <>
      <Stack spacing={2}>
        <SubCard title="通用设置">
          <Grid container spacing={{ xs: 3, sm: 2, md: 4 }}>
            <Grid xs={12}>
              <Button variant="contained" onClick={checkUpdate}>
                检查更新
              </Button>
            </Grid>
            <Grid xs={12}>
              <FormControl fullWidth>
                <TextField
                  multiline
                  maxRows={15}
                  id="Notice"
                  label="公告"
                  value={inputs.Notice}
                  name="Notice"
                  onChange={handleInputChange}
                  minRows={10}
                  placeholder="在此输入新的公告内容，支持 Markdown & HTML 代码"
                />
              </FormControl>
            </Grid>
            <Grid xs={12}>
              <Button variant="contained" onClick={submitNotice}>
                保存公告
              </Button>
            </Grid>
          </Grid>
        </SubCard>
        <SubCard title="个性化设置">
          <Grid container spacing={{ xs: 3, sm: 2, md: 4 }}>
            <Grid xs={12}>
              <FormControl fullWidth>
                <InputLabel htmlFor="SystemName">系统名称</InputLabel>
                <OutlinedInput
                  id="SystemName"
                  name="SystemName"
                  value={inputs.SystemName || ''}
                  onChange={handleInputChange}
                  label="系统名称"
                  placeholder="在此输入系统名称"
                  disabled={loading}
                />
              </FormControl>
            </Grid>
            <Grid xs={12}>
              <Button variant="contained" onClick={submitSystemName}>
                设置系统名称
              </Button>
            </Grid>
            <Grid xs={12}>
              <FormControl fullWidth>
                <InputLabel htmlFor="Theme">主题名称</InputLabel>
                <OutlinedInput
                    id="Theme"
                    name="Theme"
                    value={inputs.Theme || ''}
                    onChange={handleInputChange}
                    label="主题名称"
                    placeholder="请输入主题名称"
                    disabled={loading}
                />
              </FormControl>
            </Grid>
            <Grid xs={12}>
              <Button variant="contained" onClick={submitTheme}>
                设置主题（重启生效）
              </Button>
            </Grid>
            <Grid xs={12}>
              <FormControl fullWidth>
                <InputLabel htmlFor="Logo">Logo 图片地址</InputLabel>
                <OutlinedInput
                  id="Logo"
                  name="Logo"
                  value={inputs.Logo || ''}
                  onChange={handleInputChange}
                  label="Logo 图片地址"
                  placeholder="在此输入Logo 图片地址"
                  disabled={loading}
                />
              </FormControl>
            </Grid>
            <Grid xs={12}>
              <Button variant="contained" onClick={submitLogo}>
                设置 Logo
              </Button>
            </Grid>
            <Grid xs={12}>
              <FormControl fullWidth>
                <TextField
                  multiline
                  maxRows={15}
                  id="HomePageContent"
                  label="首页内容"
                  value={inputs.HomePageContent}
                  name="HomePageContent"
                  onChange={handleInputChange}
                  minRows={10}
                  placeholder="在此输入首页内容，支持 Markdown & HTML 代码，设置后首页的状态信息将不再显示。如果输入的是一个链接，则会使用该链接作为 iframe 的 src 属性，这允许你设置任意网页作为首页。"
                />
              </FormControl>
            </Grid>
            <Grid xs={12}>
              <Button variant="contained" onClick={() => submitOption('HomePageContent')}>
                保存首页内容
              </Button>
            </Grid>
            <Grid xs={12}>
              <FormControl fullWidth>
                <TextField
                  multiline
                  maxRows={15}
                  id="About"
                  label="关于"
                  value={inputs.About}
                  name="About"
                  onChange={handleInputChange}
                  minRows={10}
                  placeholder="在此输入新的关于内容，支持 Markdown & HTML 代码。如果输入的是一个链接，则会使用该链接作为 iframe 的 src 属性，这允许你设置任意网页作为关于页面。"
                />
              </FormControl>
            </Grid>
            <Grid xs={12}>
              <Button variant="contained" onClick={submitAbout}>
                保存关于
              </Button>
            </Grid>
            <Grid xs={12}>
              <Alert severity="warning">
                移除 One API 的版权标识必须首先获得授权，项目维护需要花费大量精力，如果本项目对你有意义，请主动支持本项目。
              </Alert>
            </Grid>
            <Grid xs={12}>
              <FormControl fullWidth>
                <TextField
                  multiline
                  maxRows={15}
                  id="Footer"
                  label="页脚"
                  value={inputs.Footer}
                  name="Footer"
                  onChange={handleInputChange}
                  minRows={10}
                  placeholder="在此输入新的页脚，留空则使用默认页脚，支持 HTML 代码"
                />
              </FormControl>
            </Grid>
            <Grid xs={12}>
              <Button variant="contained" onClick={submitFooter}>
                设置页脚
              </Button>
            </Grid>
          </Grid>
        </SubCard>
      </Stack>
      <Dialog open={showUpdateModal} onClose={() => setShowUpdateModal(false)} fullWidth maxWidth={'md'}>
        <DialogTitle sx={{ margin: '0px', fontWeight: 700, lineHeight: '1.55556', padding: '24px', fontSize: '1.125rem' }}>
          新版本：{updateData.tag_name}
        </DialogTitle>
        <Divider />
        <DialogContent>
          {' '}
          <div dangerouslySetInnerHTML={{ __html: updateData.content }}></div>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowUpdateModal(false)}>关闭</Button>
          <Button
            onClick={async () => {
              setShowUpdateModal(false);
              openGitHubRelease();
            }}
          >
            去GitHub查看
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
};

export default OtherSetting;
