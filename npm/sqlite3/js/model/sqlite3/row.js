import { Model } from '@djthorpe/js-framework';

export default class Row extends Model {
  constructor(data) {
    super({});
    this.$data = data;
  }

  static define() {
    super.define(Row, {}, 'Row');
  }
}

Row.define();
