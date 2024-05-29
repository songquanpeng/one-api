import config from 'config';
import * as actionTypes from './actions';

export const initialState = config.siteInfo;

const siteInfoReducer = (state = initialState, action) => {
  switch (action.type) {
    case actionTypes.SET_SITE_INFO:
      return {
        ...state,
        ...action.payload,
        isLoading: false,  // 添加加载状态
      };
    default:
      return state;
  }
};

export default siteInfoReducer;
