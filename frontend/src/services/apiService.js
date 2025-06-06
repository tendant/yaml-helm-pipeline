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
  
  // Get list of configuration groups
  async getConfigGroups() {
    try {
      const response = await api.get('/groups');
      return response.data;
    } catch (error) {
      console.error('Error fetching configuration groups:', error);
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

  // Preview changes for a branch and groups
  async previewChanges(branch, groups = []) {
    try {
      const response = await api.post('/preview', { branch, groups });
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

  // Commit changes to a branch with groups
  async commitChanges(branch, message, groups = []) {
    try {
      const response = await api.post('/commit', { branch, message, groups });
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
  },
  
  // Get output configuration
  async getOutputConfig() {
    try {
      const response = await api.get('/health');
      return {
        outputDir: response.data.output_dir || '/',
        outputFilename: response.data.output_filename || 'generated.yaml',
        outputRepoUrl: response.data.output_repo_url || ''
      };
    } catch (error) {
      console.error('Error fetching output config:', error);
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
