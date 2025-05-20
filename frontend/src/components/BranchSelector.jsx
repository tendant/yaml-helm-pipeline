import { For } from 'solid-js';

export function BranchSelector(props) {
  const handleChange = (e) => {
    props.onSelect(e.target.value);
  };

  return (
    <div class="mb-3">
      <label for="branch-select" class="form-label">Select Branch</label>
      <select 
        id="branch-select" 
        class="form-select" 
        value={props.selectedBranch} 
        onChange={handleChange}
        disabled={props.disabled}
      >
        <option value="" disabled>Select a branch</option>
        <For each={props.branches}>
          {(branch) => (
            <option value={branch}>{branch}</option>
          )}
        </For>
      </select>
    </div>
  );
}
