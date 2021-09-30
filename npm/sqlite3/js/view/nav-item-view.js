import { html } from 'lit-html';
import ComponentView from './component-view';
import Events from './nav-view';

customElements.define('nav-item-view', class extends ComponentView {
  constructor() {
    super();
    this.classList.add('nav-link');
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    // Fire disconnected event
    this.dispatchEvent(new CustomEvent(Events.EVENT_DISCONNECTED));
  }

  template() {
    return html`
      <link rel="stylesheet" href="./index.css">    
      <a class="${this.classList.value}" href="#">
        <slot></slot>
      </a>
    `;
  }
});
