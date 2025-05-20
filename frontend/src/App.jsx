import { createSignal, createEffect, Show } from 'solid-js';
import { BranchSelector } from './components/BranchSelector';
import { KeyPreview } from './components/KeyPreview';
import { CommitForm } from './components/CommitForm';
import { apiService } from './services/apiService';

export function App() {
  const [branches, setBranches] = createSignal([]);
  const [selectedBranch, setSelectedBranch] = createSignal('');
  const [keys, setKeys] = createSignal(null);
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal('');
  const [success, setSuccess] = createSignal('');
  const [step, setStep] = createSignal('select'); // select, preview, commit, success

  // Load branches on component mount
  createEffect(async () => {
    try {
      setLoading(true);
      setError('');
      const data = await apiService.getBranches();
      setBranches(data.branches || []);
      setLoading(false);
    } catch (err) {
      setError('Failed to load branches: ' + (err.message || 'Unknown error'));
      setLoading(false);
    }
  });

  const handleBranchSelect = (branch) => {
    setSelectedBranch(branch);
    setStep('select');
    setKeys(null);
    setError('');
    setSuccess('');
  };

  const handlePreview = async () => {
    if (!selectedBranch()) {
      setError('Please select a branch first');
      return;
    }

    try {
      setLoading(true);
      setError('');
      const data = await apiService.previewChanges(selectedBranch());
      setKeys(data.keys || {});
      setStep('preview');
      setLoading(false);
    } catch (err) {
      setError('Failed to preview changes: ' + (err.message || 'Unknown error'));
      setLoading(false);
    }
  };

  const handleCommit = async (message) => {
    if (!selectedBranch()) {
      setError('Please select a branch first');
      return;
    }

    try {
      setLoading(true);
      setError('');
      await apiService.commitChanges(selectedBranch(), message);
      setSuccess('Changes committed successfully!');
      setStep('success');
      setLoading(false);
    } catch (err) {
      setError('Failed to commit changes: ' + (err.message || 'Unknown error'));
      setLoading(false);
    }
  };

  const handleReset = () => {
    setStep('select');
    setKeys(null);
    setError('');
    setSuccess('');
  };

  return (
    <div>
      <header class="header">
        <div class="container">
          <h1>YAML Helm Pipeline</h1>
        </div>
      </header>

      <div class="container">
        <Show when={error()}>
          <div class="alert alert-danger">{error()}</div>
        </Show>

        <Show when={success()}>
          <div class="alert alert-success">{success()}</div>
        </Show>

        <div class="card">
          <div class="card-header">
            <h2 class="card-title h5 mb-0">Generate Kubernetes Secrets</h2>
          </div>
          <div class="card-body">
            <BranchSelector 
              branches={branches()} 
              selectedBranch={selectedBranch()} 
              onSelect={handleBranchSelect} 
              disabled={loading() || step() === 'commit'}
            />

            <Show when={step() === 'select'}>
              <div class="mt-3">
                <button 
                  class="btn btn-primary" 
                  onClick={handlePreview} 
                  disabled={loading() || !selectedBranch()}
                >
                  {loading() ? 'Loading...' : 'Preview Changes'}
                </button>
              </div>
            </Show>

            <Show when={step() === 'preview' && keys()}>
              <div class="mt-4">
                <h3 class="h5">Keys that will be included:</h3>
                <KeyPreview keys={keys()} />
                
                <div class="mt-3">
                  <CommitForm 
                    onCommit={handleCommit} 
                    onCancel={handleReset} 
                    disabled={loading()}
                  />
                </div>
              </div>
            </Show>

            <Show when={step() === 'success'}>
              <div class="mt-3">
                <button class="btn btn-primary" onClick={handleReset}>
                  Start New Generation
                </button>
              </div>
            </Show>
          </div>
        </div>
      </div>
    </div>
  );
}
