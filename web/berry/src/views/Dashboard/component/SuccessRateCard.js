import PropTypes from 'prop-types';

// material-ui
import { useTheme, styled } from '@mui/material/styles';
import { Box, LinearProgress, Stack, Typography } from '@mui/material';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import SkeletonTotalIncomeCard from 'ui-component/cards/Skeleton/TotalIncomeCard';

// assets
import CheckCircleIcon from '@mui/icons-material/CheckCircle';

// styles
const CardWrapper = styled(MainCard)(({ theme }) => ({
  backgroundColor: theme.palette.primary.dark,
  color: theme.palette.primary.light,
  overflow: 'hidden',
  position: 'relative',
  '&:after': {
    content: '""',
    position: 'absolute',
    width: 210,
    height: 210,
    background: `linear-gradient(210.04deg, ${theme.palette.primary[200]} -50.94%, rgba(144, 202, 249, 0) 83.49%)`,
    borderRadius: '50%',
    top: -30,
    right: -180
  },
  '&:before': {
    content: '""',
    position: 'absolute',
    width: 210,
    height: 210,
    background: `linear-gradient(140.9deg, ${theme.palette.primary[200]} -14.02%, rgba(144, 202, 249, 0) 77.58%)`,
    borderRadius: '50%',
    top: -160,
    right: -130
  }
}));

// 成功率分级配色（参考 new-api 的健康度阈值：100/90/70）
function getSuccessRateColor(rate) {
  if (rate >= 100) return '#05CD99';
  if (rate >= 90) return '#6C63FF';
  if (rate >= 70) return '#FFB547';
  return '#FF5E7D';
}

// ==============================|| DASHBOARD - SUCCESS RATE CARD ||============================== //

const SuccessRateCard = ({ isLoading, data }) => {
  const theme = useTheme();
  const total = data?.total || {};
  const success = total.success_count || 0;
  const errorCount = total.error_count || 0;
  const rate = total.success_rate || 0;
  const totalCount = success + errorCount;
  const topModels = (data?.models || []).slice(0, 5);

  return (
    <>
      {isLoading ? (
        <SkeletonTotalIncomeCard />
      ) : (
        <CardWrapper border={false} content={false}>
          <Box sx={{ p: 2, position: 'relative' }}>
            <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 1 }}>
              <CheckCircleIcon fontSize="small" />
              <Typography variant="h4">调用成功率（最近 24 小时）</Typography>
            </Stack>
            {totalCount > 0 ? (
              <>
                <Typography variant="h1" sx={{ color: getSuccessRateColor(rate), fontWeight: 700 }}>
                  {rate.toFixed(2)}%
                </Typography>
                <Typography variant="subtitle2" sx={{ color: 'primary.light', mb: 2 }}>
                  成功 {success} / 失败 {errorCount}
                </Typography>
                {topModels.length > 0 && (
                  <Box sx={{ mt: 1 }}>
                    {topModels.map((item) => (
                      <Box key={item.model_name} sx={{ mb: 1 }}>
                        <Stack direction="row" justifyContent="space-between" sx={{ mb: 0.5 }}>
                          <Typography
                            variant="body2"
                            noWrap
                            title={item.model_name}
                            sx={{ maxWidth: '60%', color: 'primary.light' }}
                          >
                            {item.model_name}
                          </Typography>
                          <Typography
                            variant="body2"
                            sx={{ color: getSuccessRateColor(item.success_rate), fontWeight: 600 }}
                          >
                            {item.success_rate.toFixed(2)}%
                          </Typography>
                        </Stack>
                        <LinearProgress
                          variant="determinate"
                          value={item.success_rate}
                          sx={{
                            height: 6,
                            borderRadius: 3,
                            backgroundColor: 'rgba(255,255,255,0.15)',
                            '& .MuiLinearProgress-bar': {
                              backgroundColor: getSuccessRateColor(item.success_rate)
                            }
                          }}
                        />
                      </Box>
                    ))}
                  </Box>
                )}
              </>
            ) : (
              <Typography variant="subtitle2" sx={{ color: 'primary.light', mt: 2 }}>
                最近 24 小时内暂无调用记录
              </Typography>
            )}
          </Box>
        </CardWrapper>
      )}
    </>
  );
};

SuccessRateCard.propTypes = {
  isLoading: PropTypes.bool,
  data: PropTypes.object
};

export default SuccessRateCard;
