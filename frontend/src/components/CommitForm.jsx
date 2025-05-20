import { createSignal } from 'solid-js';

export function CommitForm(props) {
  const [message, setMessage] = createSignal('');
  
  const handleSubmit = (e) => {
    e.preventDefault();
    if (message().trim()) {
      props.onCommit(message());
    }
  };
  
  return (
    <form onSubmit={handleSubmit}>
      <div class="mb-3">
        <label for="commit-message" class="form-label">Commit Message</label>
        <textarea
          id="commit-message"
          class="form-control"
          value={message()}
          onInput={(e) => setMessage(e.target.value)}
          placeholder="Enter a commit message"
          rows="3"
          required
          disabled={props.disabled}
        ></textarea>
      </div>
      
      <div class="d-flex gap-2">
        <button 
          type="submit" 
          class="btn btn-success" 
          disabled={props.disabled || !message().trim()}
        >
          Commit Changes
        </button>
        
        <button 
          type="button" 
          class="btn btn-secondary" 
          onClick={props.onCancel}
          disabled={props.disabled}
        >
          Cancel
        </button>
      </div>
    </form>
  );
}
