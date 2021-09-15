import { Model } from '@djthorpe/js-framework';

export default class Table extends Model {
  constructor(data) {
    super(data);
    // key is computed from the prefix
    this.$hashCode = this.name.hashCode() + this.schema.hashCode();
  }

  get key() {
    return `t-${this.$hashCode}`;
  }

  static define() {
    super.define(Table, {
      name: 'string',
      schema: 'string',
      count: 'number',
    }, 'Table');
  }
}

Table.define();
