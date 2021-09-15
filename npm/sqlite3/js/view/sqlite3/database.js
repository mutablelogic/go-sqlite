import {
  View, List, Form,
} from '@djthorpe/js-framework';

/* Constants */
const EVENT_CLICK = 'view:click';

export default class DatabaseView extends View {
  constructor(node) {
    super(node);

    // Add view for list of schemas
    const nodeSchemas = this.query('#database-schemas');
    if (nodeSchemas) {
      this.$schemas = new List(nodeSchemas, '_template');
    }

    // Add modal for list of modules
    const nodeModules = this.query('#database-modules');
    if (nodeModules) {
      this.$modules = new Form(nodeModules);
    }

    // Add list for list of modules
    const nodeModulesList = this.query('#database-modules-list');
    if (nodeModulesList) {
      this.$modulelist = new List(nodeModulesList, '_template');
    }

    // Add event listeners for buttons
    super.queryAll('.action').forEach((button) => {
      button.addEventListener('click', (evt) => {
        this.dispatchEvent(EVENT_CLICK, this, evt.srcElement);
      });
    });
  }

  /**
   * @param {Database} v
   */
  set database(v) {
    // Set version
    this.version = v.version;
    // Add schemas
    this.$schemas.clear();
    if (v.schemas) {
      v.schemas.forEach((schema) => {
        const view = this.$schemas.set(`s-${schema}`).replace('.schema', schema);
        view.$node.addEventListener('click', (evt) => {
          this.dispatchEvent(EVENT_CLICK, this, evt.srcElement);
        });
      });
    }
    // Add module list
    this.$modulelist.clear();
    if (v.modules) {
      v.modules.forEach((module) => {
        this.$modulelist.set(module).replace('.module', module);
      });
    }
    // Add pool information
    if (v.pool) {
      this.pool = v.pool;
    }
  }

  /**
   * @param {String} v
   */
  set version(v) {
    this.replace('._version', v);
  }

  /**
   * @param {Pool} pool
   */
  set pool(pool) {
    if (pool) {
      this.replace('._pool', `Pool ${pool.cur}/${pool.max}`);
    } else {
      this.replace('._pool', '');
    }
  }

  /* Show the modules modal */
  showModules() {
    this.$modules.show();
  }
}
