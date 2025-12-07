import axios from 'axios';
import type { APIResponse } from './types';

// Create axios instance with base configuration
const client = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 10000, // 10 second timeout
});

// Request interceptor (for adding auth tokens in the future)
client.interceptors.request.use(
  (config) => {
    // Future: Add authentication token here
    // config.headers.Authorization = `Bearer ${token}`;
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling
client.interceptors.response.use(
  (response) => {
    // Extract data from API envelope
    return response;
  },
  (error) => {
    if (error.response) {
      // Server responded with error
      const apiError = error.response.data as APIResponse<never>;
      console.error('API Error:', apiError.error);

      // You can add toast notifications here
      return Promise.reject(new Error(apiError.error?.message || 'Unknown error occurred'));
    } else if (error.request) {
      // Request made but no response
      console.error('Network Error:', error.message);
      return Promise.reject(new Error('Network error - please check your connection'));
    } else {
      // Something else happened
      console.error('Error:', error.message);
      return Promise.reject(error);
    }
  }
);

export default client;
