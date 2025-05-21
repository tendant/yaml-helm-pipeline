import { Show } from 'solid-js';

export function OutputInfo(props) {
  return (
    <div class="mt-3 p-3 bg-light rounded">
      <h5>Output Information</h5>
      <div>
        <strong>Repository:</strong> {props.repoUrl ? props.repoUrl : 'Same as source'}
      </div>
      <div>
        <strong>Directory:</strong> {props.dir ? props.dir : '/'}
      </div>
      <div>
        <strong>Filename:</strong> {props.filename}
      </div>
      <Show when={props.repoUrl}>
        <div class="mt-2 text-muted">
          <small>Output will be saved to a separate repository</small>
        </div>
      </Show>
    </div>
  );
}
