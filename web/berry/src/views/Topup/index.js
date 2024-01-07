import { Stack, Alert } from '@mui/material';
import Grid from '@mui/material/Unstable_Grid2';
import TopupCard from './component/TopupCard';
import InviteCard from './component/InviteCard';

const Topup = () => {
  return (
    <Grid container spacing={2}>
      <Grid xs={12}>
        <Alert severity="warning">
          充值记录以及邀请记录请在日志中查询。充值记录请在日志中选择类型【充值】查询；邀请记录请在日志中选择【系统】查询{' '}
        </Alert>
      </Grid>
      <Grid xs={12} md={6} lg={8}>
        <Stack spacing={2}>
          <TopupCard />
        </Stack>
      </Grid>
      <Grid xs={12} md={6} lg={4}>
        <InviteCard />
      </Grid>
    </Grid>
  );
};

export default Topup;
