import { Model } from '@djthorpe/js-framework';

export default class Row extends Model {
  constructor(data) {
    super({});
    console.log(data);
  }

  static define() {
    super.define(Row, {}, 'Row');
  }
}

Row.define();
