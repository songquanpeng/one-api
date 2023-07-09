import { showError } from './utils';
import axios from 'axios';

export const API = axios.create({
  baseURL: 'http://localhost:3000',
});

API.interceptors.response.use(
  (response) => response,
  (error) => {
    showError(error);
  }
);
