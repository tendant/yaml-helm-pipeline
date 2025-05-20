import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

export const apiService = {
  // Get list of branches
  async getBranches() {
    try {
      const response = await api.get('/branches');
      return response.data;
    } catch (error) {
      console.error('Error fetching branches:', error);
      throw error.response?.data || error;
    }
  },

  // Preview changes for a branch
  async previewChanges(branch) {
    try {
      const response = await api.post('/preview', { branch });
      return response.data;
    } catch (error) {
      console.error('Error previewing changes:', error);
      throw error.response?.data || error;
    }
  },

  // Commit changes to a branch
  async commitChanges(branch, message) {
    try {
      const response = await api.post('/commit', { branch, message });
      return response.data;
    } catch (error) {
      console.error('Error committing changes:', error);
      throw error.response?.data || error;
    }
  },

  // Check API health
  async checkHealth() {
    try {
      const response = await api.get('/health');
      return response.data;
    } catch (error) {
      console.error('Error checking health:', error);
      throw error.response?.data || error;
    }
  }
};
