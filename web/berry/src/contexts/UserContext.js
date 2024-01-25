// contexts/User/index.jsx
import React, { useEffect, useCallback, createContext, useState } from 'react';
import { LOGIN } from 'store/actions';
import { useDispatch } from 'react-redux';

export const UserContext = createContext();

// eslint-disable-next-line
const UserProvider = ({ children }) => {
  const dispatch = useDispatch();
  const [isUserLoaded, setIsUserLoaded] = useState(false);

  const loadUser = useCallback(() => {
    let user = localStorage.getItem('user');
    if (user) {
      let data = JSON.parse(user);
      dispatch({ type: LOGIN, payload: data });
    }
    setIsUserLoaded(true);
  }, [dispatch]);

  useEffect(() => {
    loadUser();
  }, [loadUser]);

  return <UserContext.Provider value={{ loadUser, isUserLoaded }}> {children} </UserContext.Provider>;
};

export default UserProvider;
