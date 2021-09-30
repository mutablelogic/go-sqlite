import { html } from 'lit-html';
import ComponentView from './component-view';

customElements.define('table-body-view', class extends ComponentView {
  // eslint-disable-next-line class-methods-use-this
  template() {
    return html`
      <link rel="stylesheet" href="./index.css">    
      <tbody>
        <tr><td>Row 1</td></tr>
        <tr><td>Row 2</td></tr>
        <tr><td>Row 3</td></tr>
        <tr><td>Row 4</td></tr>
      </tbody>
    `;
  }
});
