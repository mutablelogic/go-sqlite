import { Model } from '@djthorpe/js-framework';

export default class Endpoint extends Model {
  constructor(data) {
    super(data);
    // key is computed from the prefix
    this.$hashCode = this.prefix.hashCode();
  }

  get key() {
    return `v${this.$hashCode}`;
  }

  static define() {
    super.define(Endpoint, {
      prefix: 'string',
      name: 'string',
      path: 'string',
    }, 'Endpoint');
  }
}

Endpoint.define();
