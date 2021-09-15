import { Model } from '@djthorpe/js-framework';

export default class Column extends Model {
  static define() {
    super.define(Column, {
      name: 'string',
      schema: 'string',
      table: 'string',
      type: 'string',
      primary: 'boolean',
      nullable: 'boolean',
    }, 'Column');
  }
}

Column.define();
