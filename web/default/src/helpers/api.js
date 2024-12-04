import { showError } from './utils';
import { BASE_URL } from '../config';
import axios from 'axios';

export const API = axios.create({
  baseURL:
    (process.env.REACT_APP_SERVER ? process.env.REACT_APP_SERVER : '') +
    BASE_URL,
});

API.interceptors.response.use(
  (response) => response,
  (error) => {
    showError(error);
  }
);
