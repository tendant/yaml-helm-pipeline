import { createMemo, For } from 'solid-js';

export function KeyPreview(props) {
  // Convert the nested keys object to a flat list for display
  const flattenedKeys = createMemo(() => {
    const result = [];
    
    function flatten(obj, prefix = '') {
      for (const key in obj) {
        const fullPath = prefix ? `${prefix}.${key}` : key;
        
        if (typeof obj[key] === 'object' && obj[key] !== null && !Array.isArray(obj[key])) {
          flatten(obj[key], fullPath);
        } else {
          result.push({
            path: fullPath,
            value: obj[key]
          });
        }
      }
    }
    
    flatten(props.keys);
    return result;
  });

  return (
    <div class="key-preview">
      <For each={flattenedKeys()}>
        {(item) => (
          <div class="key-item">
            <span class="key-path">{item.path}: </span>
            <span class="key-value">{item.value}</span>
          </div>
        )}
      </For>
      
      {flattenedKeys().length === 0 && (
        <div class="text-muted">No keys found</div>
      )}
    </div>
  );
}
