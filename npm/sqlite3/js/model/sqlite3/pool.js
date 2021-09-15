import { Model } from '@djthorpe/js-framework';

export default class Pool extends Model {
  static define() {
    super.define(Pool, {
      cur: 'number',
      max: 'number',
    }, 'Pool');
  }
}

Pool.define();
