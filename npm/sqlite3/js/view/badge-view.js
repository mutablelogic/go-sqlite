import { html } from 'lit-html';
import ComponentView from './component-view';

// <badge-view>TEST</badge-view>
customElements.define('badge-view', class extends ComponentView {
  constructor() {
    super();
    this.classList.add('badge');
  }

  template() {
    return html`
      <link rel="stylesheet" href="./index.css">    
      <span class="${this.classList.value}">${this.textContent}</span>
    `;
  }
});
