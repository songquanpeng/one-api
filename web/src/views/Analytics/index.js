import { gridSpacing } from 'store/constant';
import { Grid } from '@mui/material';
import MainCard from 'ui-component/cards/MainCard';
import Statistics from './component/Statistics';
import Overview from './component/Overview';

export default function MarketingData() {
  return (
    <Grid container spacing={gridSpacing}>
      <Grid item xs={12}>
        <Statistics />
      </Grid>
      <Grid item xs={12}>
        <MainCard>
          <Overview />
        </MainCard>
      </Grid>
    </Grid>
  );
}
