import { createSignal, createEffect, For, Show } from 'solid-js';
import { apiService } from '../services/apiService';

export function ConfigGroupSelector(props) {
  const [groups, setGroups] = createSignal([]);
  const [selectedGroups, setSelectedGroups] = createSignal([]);
  const [loading, setLoading] = createSignal(true);
  const [error, setError] = createSignal(null);

  createEffect(async () => {
    try {
      setLoading(true);
      const data = await apiService.getConfigGroups();
      setGroups(data.groups || []);
      setError(null);
    } catch (err) {
      setError(err.message);
      setGroups([]);
    } finally {
      setLoading(false);
    }
  });

  const handleGroupChange = (group, isSelected) => {
    if (isSelected) {
      setSelectedGroups([...selectedGroups(), group]);
    } else {
      setSelectedGroups(selectedGroups().filter(g => g !== group));
    }
  };

  createEffect(() => {
    if (props.onSelect) {
      props.onSelect(selectedGroups());
    }
  });

  return (
    <div class="config-group-selector mt-3">
      <h3 class="h5">Configuration Groups</h3>
      
      <Show when={loading()}>
        <div class="loading">Loading configuration groups...</div>
      </Show>
      
      <Show when={error()}>
        <div class="alert alert-danger">Error loading configuration groups: {error()}</div>
      </Show>
      
      <Show when={!loading() && !error() && groups().length === 0}>
        <div class="alert alert-info">No configuration groups found.</div>
      </Show>
      
      <Show when={!loading() && !error() && groups().length > 0}>
        <div class="groups-list">
          <For each={groups()}>
            {(group) => (
              <div class="form-check">
                <input
                  class="form-check-input"
                  type="checkbox"
                  id={`group-${group}`}
                  checked={selectedGroups().includes(group)}
                  onChange={(e) => handleGroupChange(group, e.target.checked)}
                  disabled={props.disabled}
                />
                <label class="form-check-label" for={`group-${group}`}>
                  {group}
                </label>
              </div>
            )}
          </For>
        </div>
      </Show>
    </div>
  );
}
