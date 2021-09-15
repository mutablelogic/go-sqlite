import { Model } from '@djthorpe/js-framework';
import { Column } from './column';
import { Row } from './row';

export default class TableData extends Model {
  static define() {
    super.define(TableData, {
      schema: 'string',
      table: 'string',
      sql: 'string',
      columns: '[]Column',
      results: '[]Row',
    }, 'TableData');
  }
}

TableData.define();
