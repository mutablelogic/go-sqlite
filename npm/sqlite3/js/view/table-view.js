import { html } from 'lit-html';
import ComponentView from './component-view';
import './table-head-view';
import './table-body-view';

customElements.define('table-view', class extends ComponentView {
  constructor() {
    super();
    this.classList.add('table');
  }

  template() {
    return html`
      <link rel="stylesheet" href="./index.css">    
      <table class="${this.classList.value}">
        ${html`<table-head-view></table-head-view>`}
        ${html`<table-body-view></table-body-view>`}
      </table>
    `;
  }
});
