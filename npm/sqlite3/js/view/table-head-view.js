import { html } from 'lit-html';
import ComponentView from './component-view';

customElements.define('table-head-view', class extends ComponentView {
  // eslint-disable-next-line class-methods-use-this
  template() {
    return html`
      <link rel="stylesheet" href="./index.css">    
      <thead><tr><th>Col A</th></tr></thead>
    `;
  }
});
