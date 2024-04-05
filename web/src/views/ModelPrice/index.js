import { useState, useEffect, useCallback } from 'react';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableContainer from '@mui/material/TableContainer';
import PerfectScrollbar from 'react-perfect-scrollbar';

import { Card } from '@mui/material';
import PricesTableRow from './component/TableRow';
import TableNoData from 'ui-component/TableNoData';
import KeywordTableHead from 'ui-component/TableHead';
import { API } from 'utils/api';
import { showError } from 'utils/common';
import { ValueFormatter, priceType } from 'views/Pricing/component/util';

// ----------------------------------------------------------------------
export default function ModelPrice() {
  const [rows, setRows] = useState([]);
  const [userModelList, setUserModelList] = useState([]);
  const [prices, setPrices] = useState({});
  const [ownedby, setOwnedby] = useState([]);

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

  const fetchPrices = useCallback(async () => {
    try {
      const res = await API.get('/api/prices');
      const { success, message, data } = res.data;
      if (success) {
        let pricesObj = {};
        data.forEach((price) => {
          if (pricesObj[price.model] === undefined) {
            pricesObj[price.model] = price;
          }
        });
        setPrices(pricesObj);
      } else {
        showError(message);
      }
    } catch (error) {
      console.error(error);
    }
  }, []);

  const fetchUserModelList = useCallback(async () => {
    try {
      const res = await API.get('/api/user/models');
      if (res === undefined) {
        setUserModelList([]);
        return;
      }
      setUserModelList(res.data.data);
    } catch (error) {
      console.error(error);
    }
  }, []);

  useEffect(() => {
    if (userModelList.length === 0 || Object.keys(prices).length === 0 || ownedby.length === 0) {
      return;
    }

    let newRows = [];
    userModelList.forEach((model) => {
      const price = prices[model.id];
      const type_label = priceType.find((pt) => pt.value === price?.type);
      const channel_label = ownedby.find((ob) => ob.value === price?.channel_type);
      newRows.push({
        model: model.id,
        type: type_label?.label || '未知',
        channel_type: channel_label?.label || '未知',
        input: ValueFormatter(price?.input || 30),
        output: ValueFormatter(price?.output || 30)
      });
    });
    setRows(newRows);
  }, [userModelList, ownedby, prices]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        await Promise.all([fetchOwnedby(), fetchUserModelList()]);
        fetchPrices();
      } catch (error) {
        console.error(error);
      }
    };

    fetchData();
  }, [fetchOwnedby, fetchUserModelList, fetchPrices]);

  return (
    <>
      <Card>
        <PerfectScrollbar component="div">
          <TableContainer sx={{ overflow: 'unset' }}>
            <Table sx={{ minWidth: 800 }}>
              <KeywordTableHead
                headLabel={[
                  { id: 'model', label: '模型名称', disableSort: true },
                  { id: 'type', label: '类型', disableSort: true },
                  { id: 'channel_type', label: '供应商', disableSort: true },
                  { id: 'input', label: '输入(/1k tokens)', disableSort: true },
                  { id: 'output', label: '输出(/1k tokens)', disableSort: true }
                ]}
              />
              <TableBody>
                {rows.length === 0 ? (
                  <TableNoData message="无可用模型" />
                ) : (
                  rows.map((row) => <PricesTableRow item={row} key={row.model} />)
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </PerfectScrollbar>
      </Card>
    </>
  );
}
