import PropTypes from 'prop-types';

// material-ui
import { Grid, Typography } from '@mui/material';

// third-party
import Chart from 'react-apexcharts';

// project imports
import SkeletonTotalGrowthBarChart from 'ui-component/cards/Skeleton/TotalGrowthBarChart';
import MainCard from 'ui-component/cards/MainCard';
import { gridSpacing } from 'store/constant';
import { Box } from '@mui/material';

// ==============================|| DASHBOARD DEFAULT - TOTAL GROWTH BAR CHART ||============================== //

const ApexCharts = ({ isLoading, chartDatas, title = '统计' }) => {
  return (
    <>
      {isLoading ? (
        <SkeletonTotalGrowthBarChart />
      ) : (
        <MainCard>
          <Grid container spacing={gridSpacing}>
            <Grid item xs={12}>
              <Grid container alignItems="center" justifyContent="space-between">
                <Grid item>
                  <Typography variant="h3">{title}</Typography>
                </Grid>
              </Grid>
            </Grid>
            <Grid item xs={12}>
              {chartDatas.series ? (
                <Chart {...chartDatas} />
              ) : (
                <Box
                  sx={{
                    minHeight: '490px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center'
                  }}
                >
                  <Typography variant="h3" color={'#697586'}>
                    暂无数据
                  </Typography>
                </Box>
              )}
            </Grid>
          </Grid>
        </MainCard>
      )}
    </>
  );
};

ApexCharts.propTypes = {
  isLoading: PropTypes.bool,
  chartDatas: PropTypes.oneOfType([PropTypes.array, PropTypes.object]),
  title: PropTypes.string
};

export default ApexCharts;
