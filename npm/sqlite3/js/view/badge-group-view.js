import { html } from 'lit-html';
import ComponentView from './component-view';
import './badge-view';

customElements.define('badge-group-view', class extends ComponentView {
  // eslint-disable-next-line class-methods-use-this
  template() {
    return html`
      <badge-view class="bg-danger">TEST ONE</badge-view><badge-view class="bg-warning">TEST TWO</badge-view>
    `;
  }
});
