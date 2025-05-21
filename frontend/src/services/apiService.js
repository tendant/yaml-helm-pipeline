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
      // Extract the error message from the response
      const errorMessage = error.response?.data || 
                          (error.message || 'Unknown error');
      
      // If it's a string, use it directly, otherwise try to extract from object
      const message = typeof errorMessage === 'string' 
                      ? errorMessage 
                      : (errorMessage.message || errorMessage.error || JSON.stringify(errorMessage));
      
      throw new Error(message);
    }
  },

  // Preview changes for a branch
  async previewChanges(branch) {
    try {
      const response = await api.post('/preview', { branch });
      return response.data;
    } catch (error) {
      console.error('Error previewing changes:', error);
      // Extract the error message from the response
      const errorMessage = error.response?.data || 
                          (error.message || 'Unknown error');
      
      // If it's a string, use it directly, otherwise try to extract from object
      const message = typeof errorMessage === 'string' 
                      ? errorMessage 
                      : (errorMessage.message || errorMessage.error || JSON.stringify(errorMessage));
      
      throw new Error(message);
    }
  },

  // Commit changes to a branch
  async commitChanges(branch, message) {
    try {
      const response = await api.post('/commit', { branch, message });
      return response.data;
    } catch (error) {
      console.error('Error committing changes:', error);
      // Extract the error message from the response
      const errorMessage = error.response?.data || 
                          (error.message || 'Unknown error');
      
      // If it's a string, use it directly, otherwise try to extract from object
      const message = typeof errorMessage === 'string' 
                      ? errorMessage 
                      : (errorMessage.message || errorMessage.error || JSON.stringify(errorMessage));
      
      throw new Error(message);
    }
  },

  // Check API health
  async checkHealth() {
    try {
      const response = await api.get('/health');
      return response.data;
    } catch (error) {
      console.error('Error checking health:', error);
      // Extract the error message from the response
      const errorMessage = error.response?.data || 
                          (error.message || 'Unknown error');
      
      // If it's a string, use it directly, otherwise try to extract from object
      const message = typeof errorMessage === 'string' 
                      ? errorMessage 
                      : (errorMessage.message || errorMessage.error || JSON.stringify(errorMessage));
      
      throw new Error(message);
    }
  }
};
