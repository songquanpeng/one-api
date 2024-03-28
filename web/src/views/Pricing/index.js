import { useState, useEffect, useMemo, useCallback } from 'react';
import PropTypes from 'prop-types';
import { Tabs, Tab, Box, Card, Alert, Stack, Button } from '@mui/material';
import { IconTag, IconTags } from '@tabler/icons-react';
import Single from './single';
import Multiple from './multiple';
import { useLocation, useNavigate } from 'react-router-dom';
import AdminContainer from 'ui-component/AdminContainer';
import { API } from 'utils/api';
import { showError } from 'utils/common';
import { CheckUpdates } from './component/CheckUpdates';

function CustomTabPanel(props) {
  const { children, value, index, ...other } = props;

  return (
    <div role="tabpanel" hidden={value !== index} id={`pricing-tabpanel-${index}`} aria-labelledby={`pricing-tab-${index}`} {...other}>
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
    id: `pricing-tab-${index}`,
    'aria-controls': `pricing-tabpanel-${index}`
  };
}

const Pricing = () => {
  const [ownedby, setOwnedby] = useState([]);
  const [modelList, setModelList] = useState([]);
  const [openModal, setOpenModal] = useState(false);
  const [errPrices, setErrPrices] = useState('');
  const [prices, setPrices] = useState([]);
  const [noPriceModel, setNoPriceModel] = useState([]);

  const location = useLocation();
  const navigate = useNavigate();
  const hash = location.hash.replace('#', '');
  const tabMap = useMemo(
    () => ({
      single: 0,
      multiple: 1
    }),
    []
  );
  const [value, setValue] = useState(tabMap[hash] || 0);

  const handleChange = (event, newValue) => {
    setValue(newValue);
    const hashArray = Object.keys(tabMap);
    navigate(`#${hashArray[newValue]}`);
  };

  const reloadData = () => {
    fetchModelList();
    fetchPrices();
  };

  const handleOkModal = (status) => {
    if (status === true) {
      reloadData();
      setOpenModal(false);
    }
  };

  useEffect(() => {
    const missingModels = modelList.filter((model) => !prices.some((price) => price.model === model));
    setNoPriceModel(missingModels);
  }, [modelList, prices]);

  useEffect(() => {
    // check if there is any price that is not valid
    const invalidPrices = prices.filter((price) => price.channel_type <= 0);
    if (invalidPrices.length > 0) {
      setErrPrices(invalidPrices.map((price) => price.model).join(', '));
    } else {
      setErrPrices('');
    }
  }, [prices]);

  const fetchOwnedby = useCallback(async () => {
    try {
      const res = await API.get('/api/ownedby');
      const { success, message, data } = res.data;
      if (success) {
        let ownedbyList = [];
        for (let key in data) {
          ownedbyList.push({ value: parseInt(key), label: data[key] });
        }
        setOwnedby(ownedbyList);
      } else {
        showError(message);
      }
    } catch (error) {
      console.error(error);
    }
  }, []);

  const fetchModelList = useCallback(async () => {
    try {
      const res = await API.get('/api/prices/model_list');
      const { success, message, data } = res.data;
      if (success) {
        setModelList(data);
      } else {
        showError(message);
      }
    } catch (error) {
      console.error(error);
    }
  }, []);

  const fetchPrices = useCallback(async () => {
    try {
      const res = await API.get('/api/prices');
      const { success, message, data } = res.data;
      if (success) {
        setPrices(data);
      } else {
        showError(message);
      }
    } catch (error) {
      console.error(error);
    }
  }, []);

  useEffect(() => {
    const handleHashChange = () => {
      const hash = location.hash.replace('#', '');
      setValue(tabMap[hash] || 0);
    };
    window.addEventListener('hashchange', handleHashChange);
    return () => {
      window.removeEventListener('hashchange', handleHashChange);
    };
  }, [location, tabMap, fetchOwnedby]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        await Promise.all([fetchOwnedby(), fetchModelList()]);
        fetchPrices();
      } catch (error) {
        console.error(error);
      }
    };

    fetchData();
  }, [fetchOwnedby, fetchModelList, fetchPrices]);

  return (
    <Stack spacing={3}>
      <Alert severity="info">
        <b>美元</b>：1 === $0.002 / 1K tokens <b>人民币</b>： 1 === ￥0.014 / 1k tokens
        <br /> <b>例如</b>：<br /> gpt-4 输入： $0.03 / 1K tokens 完成：$0.06 / 1K tokens <br />
        0.03 / 0.002 = 15, 0.06 / 0.002 = 30，即输入倍率为 15，完成倍率为 30
      </Alert>

      {noPriceModel.length > 0 && (
        <Alert severity="warning">
          <b>存在未配置价格的模型，请及时配置价格</b>：
          {noPriceModel.map((model) => (
            <span key={model}>{model}, </span>
          ))}
        </Alert>
      )}

      {errPrices && (
        <Alert severity="warning">
          <b>存在供应商类型错误的模型，请及时配置</b>：{errPrices}
        </Alert>
      )}
      <Stack direction="row" alignItems="center" justifyContent="flex-end" mb={5} spacing={2}>
        <Button
          variant="contained"
          onClick={() => {
            setOpenModal(true);
          }}
        >
          更新价格
        </Button>
      </Stack>
      <Card>
        <AdminContainer>
          <Box sx={{ width: '100%' }}>
            <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
              <Tabs value={value} onChange={handleChange} variant="scrollable" scrollButtons="auto">
                <Tab label="单条操作" {...a11yProps(0)} icon={<IconTag />} iconPosition="start" />
                <Tab label="合并操作" {...a11yProps(1)} icon={<IconTags />} iconPosition="start" />
              </Tabs>
            </Box>
            <CustomTabPanel value={value} index={0}>
              <Single ownedby={ownedby} reloadData={reloadData} prices={prices} />
            </CustomTabPanel>
            <CustomTabPanel value={value} index={1}>
              <Multiple ownedby={ownedby} reloadData={reloadData} prices={prices} noPriceModel={noPriceModel} />
            </CustomTabPanel>
          </Box>
        </AdminContainer>
      </Card>
      <CheckUpdates
        open={openModal}
        onCancel={() => {
          setOpenModal(false);
        }}
        row={prices}
        onOk={handleOkModal}
      />
    </Stack>
  );
};

export default Pricing;
