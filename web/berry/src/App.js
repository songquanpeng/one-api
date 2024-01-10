import { useSelector } from 'react-redux';

import { ThemeProvider } from '@mui/material/styles';
import { CssBaseline, StyledEngineProvider } from '@mui/material';

// routing
import Routes from 'routes';

// defaultTheme
import themes from 'themes';

// project imports
import NavigationScroll from 'layout/NavigationScroll';

// auth
import UserProvider from 'contexts/UserContext';
import StatusProvider from 'contexts/StatusContext';
import { SnackbarProvider } from 'notistack';

// ==============================|| APP ||============================== //

const App = () => {
  const customization = useSelector((state) => state.customization);

  return (
    <StyledEngineProvider injectFirst>
      <ThemeProvider theme={themes(customization)}>
        <CssBaseline />
        <NavigationScroll>
          <SnackbarProvider autoHideDuration={5000} maxSnack={3} anchorOrigin={{ vertical: 'top', horizontal: 'right' }}>
            <UserProvider>
              <StatusProvider>
                <Routes />
              </StatusProvider>
            </UserProvider>
          </SnackbarProvider>
        </NavigationScroll>
      </ThemeProvider>
    </StyledEngineProvider>
  );
};

export default App;
