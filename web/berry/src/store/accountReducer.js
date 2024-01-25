import * as actionTypes from './actions';

export const initialState = {
  user: undefined
};

const accountReducer = (state = initialState, action) => {
  switch (action.type) {
    case actionTypes.LOGIN:
      return {
        ...state,
        user: action.payload
      };
    case actionTypes.LOGOUT:
      return {
        ...state,
        user: undefined
      };
    default:
      return state;
  }
};

export default accountReducer;
