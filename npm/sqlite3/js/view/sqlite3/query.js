import {
  View,
} from '@djthorpe/js-framework';

const EVENT_INPUT = 'query:input';

export default class QueryView extends View {
  constructor(node) {
    super(node);
    this.$node.addEventListener('input', () => {
      this.dispatchEvent(EVENT_INPUT, this, this.value);
    });
  }

  get value() {
    return this.$node.innerText;
  }
}
