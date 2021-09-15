import { Model } from '@djthorpe/js-framework';
import { Pool } from './pool';

export default class Database extends Model {
  static define() {
    super.define(Database, {
      version: 'string',
      modules: '[]string',
      schemas: '[]string',
      pool: 'Pool',
    }, 'Database');
  }
}

Database.define();
