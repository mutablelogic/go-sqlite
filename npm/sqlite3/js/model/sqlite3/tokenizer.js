import { Model } from '@djthorpe/js-framework';

export default class Tokenizer extends Model {
  static define() {
    super.define(Tokenizer, {
      html: '[]string',
      complete: 'boolean',
    }, 'Tokenizer');
  }
}

Tokenizer.define();
