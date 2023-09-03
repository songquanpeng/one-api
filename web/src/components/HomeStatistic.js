import React, { useEffect, useRef, useState } from 'react';
import CountUp from "react-countup";
import { Segment, Statistic } from 'semantic-ui-react';
import { API, showError } from '../helpers';
import { usePrevious } from 'ahooks';

const HomeStatistic = () => {

  const [totalQuota, setTotalQuota] = useState();
  const [requestCount, setRequestCount] = useState('');
  const [todayQuota, setTodayQuota] = useState('');
  const [todayCount, setTodayCount] = useState('');
  const [unitCount, setUnitCount] = useState('');

  const oldTotalQuota = usePrevious(totalQuota);
  const oldRequestCount = usePrevious(requestCount);
  const oldTodayQuota = usePrevious(todayQuota);
  const oldTodayCount = usePrevious(todayCount);


  const getServerCount = async () => {
    const res = await API.get('/api/server_count');
    const { success, message, data } = res.data;
    if (success) {
      setTotalQuota(data.totalQuota)
      if (data.requestCount / 1000000 >= 10) {
        setRequestCount(data.requestCount / 1000000)
        setUnitCount('m')
      } else {
        setRequestCount(data.requestCount)
      }

      setTodayQuota(data.todayQuota)
      setTodayCount(data.todayCount)
    } else {
      showError(message);
    }
  }

  const timer = useRef(null)
  useEffect(() => {
    if (!timer.current) {
      getServerCount();
    }
    timer.current = setInterval(function () {
      getServerCount();
    }, 15000);
    return () => {
      if (timer.current) {
        clearInterval(timer.current)
      }
    }
  }, []);

  return (
    <Segment inverted className="statistic-box">
      <Statistic color='red' inverted size='large'>
        <Statistic.Value>$<CountUp start={oldTotalQuota} end={totalQuota} duration="3" redraw="true" /></Statistic.Value>
        <Statistic.Label>当前消耗总量</Statistic.Label>
      </Statistic>
      <Statistic color='red' inverted size='large'>
        <Statistic.Value><CountUp start={oldRequestCount} end={requestCount} duration="3" redraw="true" />{unitCount}</Statistic.Value>
        <Statistic.Label>当前调用次数</Statistic.Label>
      </Statistic>
      <Statistic color='red' inverted size='large'>
        <Statistic.Value>$<CountUp start={oldTodayQuota} end={todayQuota} duration="3" redraw="true" /></Statistic.Value>
        <Statistic.Label>今日消耗总量</Statistic.Label>
      </Statistic>
      <Statistic color='red' inverted size='large'>
        <Statistic.Value><CountUp start={oldTodayCount} end={todayCount} duration="3" redraw="true" /></Statistic.Value>
        <Statistic.Label>今日调用次数</Statistic.Label>
      </Statistic>
    </Segment>
  );
};

export default HomeStatistic;
