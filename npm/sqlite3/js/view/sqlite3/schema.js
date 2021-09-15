import {
  View, List,
} from '@djthorpe/js-framework';
import Component from '../component';

// Constants
const EVENT_CHANGE = 'schema:change';
const EVENT_CLICK = 'schema:click';

export default class SchemaView extends View {
  constructor(node) {
    super(node);

    // Add view for list of tables (tabs at the top of the page)
    const nodeTableList = this.query('#schema-table-list');
    if (nodeTableList) {
      this.$tables = new List(nodeTableList, '_template');
    }
  }

  /**
   * @param {Schema} v
   */
  set schema(v) {
    // Set name, filename and memory badge
    this.name = v.schema;
    if (v.filename) {
      this.filename = v.filename;
    } else {
      this.memory = v.memory;
    }

    // Get list of names already in the table
    const tabMap = new Map();
    let changed = false;
    this.$tables.queryAll('._name').forEach((node) => {
      tabMap.set(node.textContent, node.parentNode);
    });

    // Enumerate tables, remove any added to the map
    if (v.tables) {
      v.tables.forEach((table) => {
        const view = this.$tables.set(`${table.key}`).replace('._name', table.name);
        if (!tabMap.has(table.name)) {
          changed = true;
          view.$node.addEventListener('click', () => {
            this.dispatchEvent(EVENT_CLICK, this, table);
          });
        }
        tabMap.delete(table.name);
      });
    }

    // Remove any nodes not in the map
    tabMap.forEach((node) => {
      changed = true;
      node.parentNode.removeChild(node);
    });

    // Fire changed event
    if (changed) {
      this.dispatchEvent(EVENT_CHANGE, this);
    }
  }

  /**
   * @param {String} v
   */
  set name(v) {
    this.replace('._name', v);
  }

  /**
   * @param {String} v
   */
  set filename(v) {
    if (v) {
      this.replace('._filename', v);
    } else {
      this.replace('._filename', '');
    }
  }

  /**
   * @param {boolean} v
   */
  set memory(v) {
    if (v) {
      this.replace('._filename', Component.badge('bg-secondary', 'MEMORY'));
    } else {
      this.replace('._filename', '');
    }
  }

  /**
   * @param {Table} table
   */
  set active(table) {
    // Remove all active classes, then add active class
    this.$tables.queryAll('.nav-item .nav-link').forEach((node) => {
      node.classList.remove('active');
    });
    const row = this.$tables.getForKey(`${table.key}`);
    if (row) {
      row.querySelector('.nav-link').classList.add('active');
    }
  }
}
