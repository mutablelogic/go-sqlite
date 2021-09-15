import { Model } from '@djthorpe/js-framework';

export default class Database extends Model {
  static define() {
    super.define(Database, {
      version: 'string',
      modules: '[]string',
      schemas: '[]string',
    }, 'Database');
  }
}

Database.define();
