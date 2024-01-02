import { Outlet } from 'react-router-dom';
import { useTheme } from '@mui/material/styles';
import { AppBar, Box, CssBaseline, Toolbar } from '@mui/material';
import Header from './Header';
import Footer from 'ui-component/Footer';

// ==============================|| MINIMAL LAYOUT ||============================== //

const MinimalLayout = () => {
  const theme = useTheme();

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <CssBaseline />
      <AppBar
        enableColorOnDark
        position="fixed"
        color="inherit"
        elevation={0}
        sx={{
          bgcolor: theme.palette.background.default,
          flex: 'none'
        }}
      >
        <Toolbar>
          <Header />
        </Toolbar>
      </AppBar>
      <Box sx={{ flex: '1 1 auto', overflow: 'auto' }} paddingTop={'64px'}>
        <Outlet />
      </Box>
      <Box sx={{ flex: 'none' }}>
        <Footer />
      </Box>
    </Box>
  );
};

export default MinimalLayout;
