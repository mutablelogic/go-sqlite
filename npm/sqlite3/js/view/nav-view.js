import { html } from 'lit-html';
import ComponentView from './component-view';

const EVENT_CLICK = 'nav-view:click';
const EVENT_ACTIVE = 'nav-view:active';
const EVENT_DEACTIVE = 'nav-view:deactive';
const EVENT_DISCONNECTED = 'nav-view:disconnected';

customElements.define('nav-view', class extends ComponentView {
  constructor() {
    super();
    this.classList.add('nav');
  }

  get active() {
    return this.querySelector('nav-item-view#active');
  }

  set active(activeItem) {
    this.querySelectorAll('nav-item-view').forEach((item) => {
      if (activeItem === item) {
        if (!item.classList.contains('active')) {
          item.classList.add('active');
          this.dispatchEvent(new CustomEvent(EVENT_ACTIVE, { detail: item }));
        }
      } else if (item.classList.contains('active')) {
        item.classList.remove('active');
        this.dispatchEvent(new CustomEvent(EVENT_DEACTIVE, { detail: item }));
      }
    });
  }

  connectedCallback() {
    super.connectedCallback();

    this.querySelectorAll('nav-item-view').forEach((item) => {
      // Add event listeners for 'click' event to nav-item-view elements
      item.addEventListener('click', (e) => {
        this.dispatchEvent(new CustomEvent(EVENT_CLICK, {
          detail: e.target,
        }));
      });
      // Add event listeners for when an item is removed from the nav-view
      item.addEventListener(EVENT_DISCONNECTED, (e) => {
        if (e.target === this.active) {
          this.active = null;
        }
      });
    });

    // If no active item is set, set the first one as active
    if (!this.active) {
      this.active = this.querySelector('nav-item-view');
    }
  }

  appendChild(element) {
    super.appendChild(element);
    this.update();
    return element;
  }

  template() {
    return html`
      <link rel="stylesheet" href="./index.css">    
      <nav class="${this.classList.value}">
        <slot></slot>
      </nav>
    `;
  }
});

// Export event names
export default {
  EVENT_CLICK, EVENT_ACTIVE, EVENT_DEACTIVE, EVENT_DISCONNECTED,
};
