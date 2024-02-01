import { useState, useEffect, useCallback } from 'react';
import { Grid } from '@mui/material';
import DataCard from 'ui-component/cards/DataCard';
import { gridSpacing } from 'store/constant';
import { showError, renderQuota } from 'utils/common';
import { API } from 'utils/api';

export default function Overview() {
  const [userLoading, setUserLoading] = useState(true);
  const [channelLoading, setChannelLoading] = useState(true);
  const [redemptionLoading, setRedemptionLoading] = useState(true);
  const [userStatistics, setUserStatistics] = useState({});

  const [channelStatistics, setChannelStatistics] = useState({
    active: 0,
    disabled: 0,
    test_disabled: 0,
    total: 0
  });
  const [redemptionStatistics, setRedemptionStatistics] = useState({
    total: 0,
    used: 0,
    unused: 0
  });

  const userStatisticsData = useCallback(async () => {
    try {
      const res = await API.get('/api/analytics/user_statistics');
      const { success, message, data } = res.data;
      if (success) {
        data.total_quota = renderQuota(data.total_quota);
        data.total_used_quota = renderQuota(data.total_used_quota);
        data.total_direct_user = data.total_user - data.total_inviter_user;
        setUserStatistics(data);
        setUserLoading(false);
      } else {
        showError(message);
      }
    } catch (error) {
      return;
    }
  }, []);

  const channelStatisticsData = useCallback(async () => {
    try {
      const res = await API.get('/api/analytics/channel_statistics');
      const { success, message, data } = res.data;
      if (success) {
        let channelData = channelStatistics;
        channelData.total = 0;
        data.forEach((item) => {
          if (item.status === 1) {
            channelData.active = item.total_channels;
          } else if (item.status === 2) {
            channelData.disabled = item.total_channels;
          } else if (item.status === 3) {
            channelData.test_disabled = item.total_channels;
          }
          channelData.total += item.total_channels;
        });
        setChannelStatistics(channelData);
        setChannelLoading(false);
      } else {
        showError(message);
      }
    } catch (error) {
      return;
    }
  }, [channelStatistics]);

  const redemptionStatisticsData = useCallback(async () => {
    try {
      const res = await API.get('/api/analytics/redemption_statistics');
      const { success, message, data } = res.data;
      if (success) {
        let redemptionData = redemptionStatistics;
        redemptionData.total = 0;
        data.forEach((item) => {
          if (item.status === 1) {
            redemptionData.unused = renderQuota(item.quota);
          } else if (item.status === 3) {
            redemptionData.used = renderQuota(item.quota);
          }
          redemptionData.total += item.quota;
        });
        redemptionData.total = renderQuota(redemptionData.total);
        setRedemptionStatistics(redemptionData);
        setRedemptionLoading(false);
      } else {
        showError(message);
      }
    } catch (error) {
      return;
    }
  }, [redemptionStatistics]);

  useEffect(() => {
    userStatisticsData();
    channelStatisticsData();
    redemptionStatisticsData();
  }, [userStatisticsData, channelStatisticsData, redemptionStatisticsData]);

  return (
    <Grid container spacing={gridSpacing}>
      <Grid item lg={3} xs={12}>
        <DataCard
          isLoading={userLoading}
          title="用户总消费金额"
          content={userStatistics?.total_used_quota || '0'}
          subContent={'用户总余额：' + (userStatistics?.total_quota || '0')}
        />
      </Grid>
      <Grid item lg={3} xs={12}>
        <DataCard
          isLoading={userLoading}
          title="用户总数"
          content={userStatistics?.total_user || '0'}
          subContent={
            <>
              直接注册：{userStatistics?.total_direct_user || '0'} <br /> 邀请注册：{userStatistics?.total_inviter_user || '0'}
            </>
          }
        />
      </Grid>
      <Grid item lg={3} xs={12}>
        <DataCard
          isLoading={channelLoading}
          title="渠道数量"
          content={channelStatistics.total}
          subContent={
            <>
              正常：{channelStatistics.active} / 禁用：{channelStatistics.disabled} / 测试禁用：{channelStatistics.test_disabled}
            </>
          }
        />
      </Grid>
      <Grid item lg={3} xs={12}>
        <DataCard
          isLoading={redemptionLoading}
          title="兑换码发行量"
          content={redemptionStatistics.total}
          subContent={
            <>
              已使用: {redemptionStatistics.used} <br /> 未使用: {redemptionStatistics.unused}
            </>
          }
        />
      </Grid>
    </Grid>
  );
}
