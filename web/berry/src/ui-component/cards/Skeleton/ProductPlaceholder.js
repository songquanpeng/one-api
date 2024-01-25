// material-ui
import { CardContent, Grid, Skeleton, Stack } from '@mui/material';

// project import
import MainCard from '../MainCard';

// ===========================|| SKELETON TOTAL GROWTH BAR CHART ||=========================== //

const ProductPlaceholder = () => (
  <MainCard content={false} boxShadow>
    <Skeleton variant="rectangular" height={220} />
    <CardContent sx={{ p: 2 }}>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Skeleton variant="rectangular" height={20} />
        </Grid>
        <Grid item xs={12}>
          <Skeleton variant="rectangular" height={45} />
        </Grid>
        <Grid item xs={12} sx={{ pt: '8px !important' }}>
          <Stack direction="row" alignItems="center" spacing={1}>
            <Skeleton variant="rectangular" height={20} width={90} />
            <Skeleton variant="rectangular" height={20} width={38} />
          </Stack>
        </Grid>
        <Grid item xs={12}>
          <Stack direction="row" justifyContent="space-between" alignItems="center">
            <Grid container spacing={1}>
              <Grid item>
                <Skeleton variant="rectangular" height={20} width={40} />
              </Grid>
              <Grid item>
                <Skeleton variant="rectangular" height={17} width={20} />
              </Grid>
            </Grid>
            <Skeleton variant="rectangular" height={32} width={47} />
          </Stack>
        </Grid>
      </Grid>
    </CardContent>
  </MainCard>
);

export default ProductPlaceholder;
