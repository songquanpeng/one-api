// material-ui
import Skeleton from '@mui/material/Skeleton';
import SubCard from 'ui-component/cards/SubCard';
import { Divider, Stack } from '@mui/material';

const DataCard = () => (
  <SubCard sx={{ height: '160px' }}>
    <Stack spacing={1}>
      <Skeleton variant="rectangular" height={20} width={80} />
      <Skeleton variant="rectangular" height={41} width={50} />
      <Divider />
      <Skeleton variant="rectangular" />
    </Stack>
  </SubCard>
);

export default DataCard;
