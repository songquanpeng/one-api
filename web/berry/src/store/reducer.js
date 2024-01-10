import { combineReducers } from 'redux';

// reducer import
import customizationReducer from './customizationReducer';
import accountReducer from './accountReducer';
import siteInfoReducer from './siteInfoReducer';

// ==============================|| COMBINE REDUCER ||============================== //

const reducer = combineReducers({
  customization: customizationReducer,
  account: accountReducer,
  siteInfo: siteInfoReducer
});

export default reducer;
