// material-ui
import { styled } from '@mui/material/styles';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';
import { useEffect, useContext } from 'react';
import { UserContext } from 'contexts/UserContext';

// ==============================|| AUTHENTICATION 1 WRAPPER ||============================== //

const AuthStyle = styled('div')(({ theme }) => ({
  backgroundColor: theme.palette.background.default
}));

// eslint-disable-next-line
const AuthWrapper = ({ children }) => {
  const account = useSelector((state) => state.account);
  const { isUserLoaded } = useContext(UserContext);
  const navigate = useNavigate();
  useEffect(() => {
    if (isUserLoaded && account.user) {
      navigate('/panel');
    }
  }, [account, navigate, isUserLoaded]);

  return <AuthStyle> {children} </AuthStyle>;
};

export default AuthWrapper;
