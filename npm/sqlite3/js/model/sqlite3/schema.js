import { Model } from '@djthorpe/js-framework';
import { Table } from './table';

export default class Schema extends Model {
  static define() {
    super.define(Schema, {
      schema: 'string',
      filename: 'string',
      memory: 'boolean',
      tables: '[]Table',
    }, 'Schema');
  }
}

Schema.define();
