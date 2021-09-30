import { render, nothing } from 'lit-html';

export default class ComponentView extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: 'open' });
  }

  connectedCallback() {
    this.update();
  }

  // eslint-disable-next-line class-methods-use-this
  disconnectedCallback() {}

  // eslint-disable-next-line class-methods-use-this
  template() {
    return nothing;
  }

  update() {
    render(this.template(), this.shadowRoot, { eventContext: this });
  }
}
