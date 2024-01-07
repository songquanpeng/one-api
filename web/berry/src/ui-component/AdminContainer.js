import { styled } from '@mui/material/styles';
import { Container } from '@mui/material';

const AdminContainer = styled(Container)(({ theme }) => ({
  [theme.breakpoints.down('md')]: {
    paddingLeft: '0px',
    paddingRight: '0px'
  }
}));

export default AdminContainer;
