import { useState, useEffect, useCallback } from 'react';
import { Grid, Typography, Divider } from '@mui/material';
import { gridSpacing } from 'store/constant';
import DateRangePicker from 'ui-component/DateRangePicker';
import ApexCharts from 'ui-component/chart/ApexCharts';
import { showError, calculateQuota } from 'utils/common';
import dayjs from 'dayjs';
import { API } from 'utils/api';
import { generateBarChartOptions, renderChartNumber } from 'utils/chart';

export default function Overview() {
  const [channelLoading, setChannelLoading] = useState(true);
  const [redemptionLoading, setRedemptionLoading] = useState(true);
  const [usersLoading, setUsersLoading] = useState(true);
  const [channelData, setChannelData] = useState([]);
  const [redemptionData, setRedemptionData] = useState([]);
  const [usersData, setUsersData] = useState([]);
  const [dateRange, setDateRange] = useState({ start: dayjs().subtract(6, 'day').startOf('day'), end: dayjs().endOf('day') });
  const handleDateRangeChange = (value) => {
    setDateRange(value);
  };

  const channelChart = useCallback(async () => {
    setChannelLoading(true);
    try {
      const res = await API.get('/api/analytics/channel_period', {
        params: {
          start_timestamp: dateRange.start.unix(),
          end_timestamp: dateRange.end.unix()
        }
      });
      const { success, message, data } = res.data;
      if (success) {
        if (data) {
          setChannelData(getBarChartOptions(data, dateRange));
        }
      } else {
        showError(message);
      }
      setChannelLoading(false);
    } catch (error) {
      return;
    }
  }, [dateRange]);

  const redemptionChart = useCallback(async () => {
    setRedemptionLoading(true);
    try {
      const res = await API.get('/api/analytics/redemption_period', {
        params: {
          start_timestamp: dateRange.start.unix(),
          end_timestamp: dateRange.end.unix()
        }
      });
      const { success, message, data } = res.data;
      if (success) {
        if (data) {
          let chartData = getRedemptionData(data, dateRange);
          setRedemptionData(chartData);
        }
      } else {
        showError(message);
      }
      setRedemptionLoading(false);
    } catch (error) {
      return;
    }
  }, [dateRange]);

  const usersChart = useCallback(async () => {
    setUsersLoading(true);
    try {
      const res = await API.get('/api/analytics/users_period', {
        params: {
          start_timestamp: dateRange.start.unix(),
          end_timestamp: dateRange.end.unix()
        }
      });
      const { success, message, data } = res.data;
      if (success) {
        if (data) {
          setUsersData(getUsersData(data, dateRange));
        }
      } else {
        showError(message);
      }
      setUsersLoading(false);
    } catch (error) {
      return;
    }
  }, [dateRange]);

  useEffect(() => {
    channelChart();
    redemptionChart();
    usersChart();
  }, [dateRange, channelChart, redemptionChart, usersChart]);

  return (
    <Grid container spacing={gridSpacing}>
      <Grid item lg={8} xs={12}>
        <DateRangePicker defaultValue={dateRange} onChange={handleDateRangeChange} localeText={{ start: '开始时间', end: '结束时间' }} />
      </Grid>
      <Grid item xs={12}>
        <Typography variant="h3">
          {dateRange.start.format('YYYY-MM-DD')} - {dateRange.end.format('YYYY-MM-DD')}
        </Typography>
      </Grid>
      <Grid item xs={12}>
        <Divider />
      </Grid>
      <Grid item xs={12}>
        <ApexCharts id="cost" isLoading={channelLoading} chartDatas={channelData?.costs || {}} title="消费统计" decimal={3} />
      </Grid>
      <Grid item xs={12}>
        <ApexCharts id="token" isLoading={channelLoading} chartDatas={channelData?.tokens || {}} title="Tokens统计" unit="" />
      </Grid>
      <Grid item xs={12}>
        <ApexCharts id="latency" isLoading={channelLoading} chartDatas={channelData?.latency || {}} title="平均延迟" unit="" />
      </Grid>
      <Grid item xs={12}>
        <ApexCharts id="requests" isLoading={channelLoading} chartDatas={channelData?.requests || {}} title="请求数" unit="" />
      </Grid>
      <Grid item xs={12}>
        <ApexCharts isLoading={redemptionLoading} chartDatas={redemptionData} title="兑换统计" />
      </Grid>
      <Grid item xs={12}>
        <ApexCharts isLoading={usersLoading} chartDatas={usersData} title="注册统计" />
      </Grid>
    </Grid>
  );
}

function getDates(start, end) {
  var dates = [];
  var current = start;

  while (current.isBefore(end) || current.isSame(end)) {
    dates.push(current.format('YYYY-MM-DD'));
    current = current.add(1, 'day');
  }

  return dates;
}

function calculateDailyData(item, dateMap) {
  const index = dateMap.get(item.Date);
  if (index === undefined) return null;

  return {
    name: item.Channel,
    costs: calculateQuota(item.Quota, 3),
    tokens: item.PromptTokens + item.CompletionTokens,
    requests: item.RequestCount,
    latency: Number(item.RequestTime / 1000 / item.RequestCount).toFixed(3),
    index: index
  };
}

function getBarDataGroup(data, dates) {
  const dateMap = new Map(dates.map((date, index) => [date, index]));

  const result = {
    costs: { total: 0, data: new Map() },
    tokens: { total: 0, data: new Map() },
    requests: { total: 0, data: new Map() },
    latency: { total: 0, data: new Map() }
  };

  for (const item of data) {
    const dailyData = calculateDailyData(item, dateMap);
    if (!dailyData) continue;

    for (let key in result) {
      if (!result[key].data.has(dailyData.name)) {
        result[key].data.set(dailyData.name, { name: dailyData.name, data: new Array(dates.length).fill(0) });
      }
      const channelDailyData = result[key].data.get(dailyData.name);
      channelDailyData.data[dailyData.index] = dailyData[key];
      result[key].total += Number(dailyData[key]);
    }
  }
  return result;
}

function getBarChartOptions(data, dateRange) {
  const dates = getDates(dateRange.start, dateRange.end);
  const result = getBarDataGroup(data, dates);

  let channelData = {};

  channelData.costs = generateBarChartOptions(dates, Array.from(result.costs.data.values()), '美元', 3);
  channelData.costs.options.title.text = '总消费：$' + renderChartNumber(result.costs.total, 3);

  channelData.tokens = generateBarChartOptions(dates, Array.from(result.tokens.data.values()), '', 0);
  channelData.tokens.options.title.text = '总Tokens：' + renderChartNumber(result.tokens.total, 0);

  channelData.requests = generateBarChartOptions(dates, Array.from(result.requests.data.values()), '次', 0);
  channelData.requests.options.title.text = '总请求数：' + renderChartNumber(result.requests.total, 0);

  // 获取每天所有渠道的平均延迟
  let latency = Array.from(result.latency.data.values());
  let sums = [];
  let counts = [];
  for (let obj of latency) {
    for (let i = 0; i < obj.data.length; i++) {
      let value = parseFloat(obj.data[i]);
      sums[i] = sums[i] || 0;
      counts[i] = counts[i] || 0;
      if (value !== 0) {
        sums[i] = (sums[i] || 0) + value;
        counts[i] = (counts[i] || 0) + 1;
      }
    }
  }

  // 追加latency列表后面
  latency[latency.length] = {
    name: '平均延迟',
    data: sums.map((sum, i) => Number(counts[i] ? sum / counts[i] : 0).toFixed(3))
  };

  let dashArray = new Array(latency.length - 1).fill(0);
  dashArray.push(5);

  channelData.latency = generateBarChartOptions(dates, latency, '秒', 3);
  channelData.latency.type = 'line';
  channelData.latency.options.chart = {
    type: 'line',
    zoom: {
      enabled: false
    },
    background: 'transparent'
  };
  channelData.latency.options.stroke = {
    curve: 'smooth',
    dashArray: dashArray
  };

  return channelData;
}

function getRedemptionData(data, dateRange) {
  const dates = getDates(dateRange.start, dateRange.end);
  const result = [
    {
      name: '兑换金额($)',
      type: 'column',
      data: new Array(dates.length).fill(0)
    },
    {
      name: '独立用户(人)',
      type: 'line',
      data: new Array(dates.length).fill(0)
    }
  ];

  for (const item of data) {
    const index = dates.indexOf(item.date);
    if (index !== -1) {
      result[0].data[index] = calculateQuota(item.quota, 3);
      result[1].data[index] = item.user_count;
    }
  }

  let chartData = {
    height: 480,
    options: {
      chart: {
        type: 'line',
        background: 'transparent'
      },
      stroke: {
        width: [0, 4]
      },
      dataLabels: {
        enabled: true,
        enabledOnSeries: [1]
      },
      xaxis: {
        type: 'category',
        categories: dates
      },
      yaxis: [
        {
          title: {
            text: '兑换金额($)'
          }
        },
        {
          opposite: true,
          title: {
            text: '独立用户(人)'
          }
        }
      ],
      tooltip: {
        theme: 'dark'
      }
    },
    series: result
  };

  return chartData;
}

function getUsersData(data, dateRange) {
  const dates = getDates(dateRange.start, dateRange.end);
  const result = [
    {
      name: '直接注册',
      data: new Array(dates.length).fill(0)
    },
    {
      name: '邀请注册',
      data: new Array(dates.length).fill(0)
    }
  ];

  let total = 0;

  for (const item of data) {
    const index = dates.indexOf(item.date);
    if (index !== -1) {
      result[0].data[index] = item.user_count - item.inviter_user_count;
      result[1].data[index] = item.inviter_user_count;

      total += item.user_count;
    }
  }

  let chartData = generateBarChartOptions(dates, result, '人', 0);
  chartData.options.title.text = '总注册人数：' + total;

  return chartData;
}
