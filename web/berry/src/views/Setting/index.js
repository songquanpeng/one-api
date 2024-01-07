import { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { Tabs, Tab, Box, Card } from '@mui/material';
import { IconSettings2, IconActivity, IconSettings } from '@tabler/icons-react';
import OperationSetting from './component/OperationSetting';
import SystemSetting from './component/SystemSetting';
import OtherSetting from './component/OtherSetting';
import AdminContainer from 'ui-component/AdminContainer';
import { useLocation, useNavigate } from 'react-router-dom';

function CustomTabPanel(props) {
  const { children, value, index, ...other } = props;

  return (
    <div role="tabpanel" hidden={value !== index} id={`setting-tabpanel-${index}`} aria-labelledby={`setting-tab-${index}`} {...other}>
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

CustomTabPanel.propTypes = {
  children: PropTypes.node,
  index: PropTypes.number.isRequired,
  value: PropTypes.number.isRequired
};

function a11yProps(index) {
  return {
    id: `setting-tab-${index}`,
    'aria-controls': `setting-tabpanel-${index}`
  };
}

const Setting = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const hash = location.hash.replace('#', '');
  const tabMap = {
    operation: 0,
    system: 1,
    other: 2
  };
  const [value, setValue] = useState(tabMap[hash] || 0);

  const handleChange = (event, newValue) => {
    setValue(newValue);
    const hashArray = Object.keys(tabMap);
    navigate(`#${hashArray[newValue]}`);
  };

  useEffect(() => {
    const handleHashChange = () => {
      const hash = location.hash.replace('#', '');
      setValue(tabMap[hash] || 0);
    };
    window.addEventListener('hashchange', handleHashChange);
    return () => {
      window.removeEventListener('hashchange', handleHashChange);
    };
  }, [location, tabMap]);

  return (
    <>
      <Card>
        <AdminContainer>
          <Box sx={{ width: '100%' }}>
            <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
              <Tabs value={value} onChange={handleChange} variant="scrollable" scrollButtons="auto">
                <Tab label="运营设置" {...a11yProps(0)} icon={<IconActivity />} iconPosition="start" />
                <Tab label="系统设置" {...a11yProps(1)} icon={<IconSettings />} iconPosition="start" />
                <Tab label="其他设置" {...a11yProps(2)} icon={<IconSettings2 />} iconPosition="start" />
              </Tabs>
            </Box>
            <CustomTabPanel value={value} index={0}>
              <OperationSetting />
            </CustomTabPanel>
            <CustomTabPanel value={value} index={1}>
              <SystemSetting />
            </CustomTabPanel>
            <CustomTabPanel value={value} index={2}>
              <OtherSetting />
            </CustomTabPanel>
          </Box>
        </AdminContainer>
      </Card>
    </>
  );
};

export default Setting;
