import {
  Controller, Nav, Toast, Provider,
} from '@djthorpe/js-framework';

// Models
import Endpoint from '../model/static/endpoint';
import Database from '../model/sqlite3/database';
import Schema from '../model/sqlite3/schema';
import TableData from '../model/sqlite3/tabledata';

// Views
import DatabaseView from '../view/sqlite3/database';
import SchemaView from '../view/sqlite3/schema';

// Constants
const API_STATIC_PREFIX = '/api/static';
const API_SQLITE_PREFIX = '/api/sqlite';
const API_SQLITE_DELTA = 30 * 1000;
const API_STATIC_DELTA = 30 * 1000;
const DEFAULT_SCHEMA = 'main';

export default class App extends Controller {
  constructor() {
    super();

    // VIEWS
    const navNode = document.querySelector('#nav');
    if (navNode) {
      super.define('nav', new Nav(navNode));
    }

    const toastNode = document.querySelector('#toast');
    if (toastNode) {
      super.define('toast', new Toast(toastNode));
    }

    const databaseNode = document.querySelector('#database');
    super.define('databaseview', new DatabaseView(databaseNode));
    this.databaseview.addEventListener(['view:click'], (sender, target) => {
      if (target.classList.contains('action-modules')) {
        this.databaseview.showModules();
      } else if (target.classList.contains('schema')) {
        const schemaName = target.innerText;
        this.schema.request(`/${schemaName}`, null, API_SQLITE_DELTA);
      } else {
        console.log('view:click', target);
      }
    });

    const schemaNode = document.querySelector('#schema');
    super.define('schemaview', new SchemaView(schemaNode));
    this.schemaview.addEventListener(['schema:change'], () => {
      console.log('schema:change');

      // Set active table to the first table in the schema
    });
    this.schemaview.addEventListener(['schema:click'], (sender, target) => {
      console.log(`schema:click ${target}`);

      // Set active
      this.schemaview.active = target;

      // Load the table data
      this.tabledata.do(`/${target.schema}/${target.name}`);
    });

    // PROVIDERS
    super.define('static', new Provider(Endpoint, API_STATIC_PREFIX));
    if (this.static) {
      this.static.addEventListener('provider:error', (sender, error) => {
        this.toast.show(error);
      });
      this.static.addEventListener(['provider:added', 'provider:changed'], (sender, endpoint) => {
        console.log(`endpoint added or changed: ${endpoint}`);
      });
      this.static.addEventListener('provider:deleted', (sender, endpoint) => {
        console.log(`endpoint deleted: ${endpoint}`);
      });
    }

    super.define('sqlite', new Provider(Database, API_SQLITE_PREFIX));
    if (this.sqlite) {
      this.sqlite.addEventListener('provider:error', (sender, error) => {
        this.toast.show(error);
      });
      this.sqlite.addEventListener(['provider:added'], (sender, database) => {
        console.log(`sqlite added: ${database}`);

        // Set the database view
        this.databaseview.database = database;

        // Load the 'main' schema
        this.schema.request(`/${DEFAULT_SCHEMA}`, null, API_SQLITE_DELTA);
      });
      this.sqlite.addEventListener(['provider:changed'], (sender, database) => {
        console.log(`sqlite changed: ${database}`);
        this.databaseview.database = database;
      });
      this.sqlite.addEventListener('provider:deleted', (sender, database) => {
        console.log(`sqlite deleted: ${database}`);
      });
    }

    super.define('schema', new Provider(Schema, API_SQLITE_PREFIX));
    this.schema.addEventListener('provider:error', (sender, error) => {
      this.toast.show(error);
    });
    this.schema.addEventListener(['provider:added', 'provider:changed'], (sender, schema) => {
      console.log(`schema added: ${schema}`);
      this.schemaview.schema = schema;
    });
    this.schema.addEventListener('provider:deleted', (sender, schema) => {
      console.log(`schema deleted: ${schema}`);
    });

    super.define('tabledata', new Provider(TableData, API_SQLITE_PREFIX));
    this.tabledata.addEventListener('provider:error', (sender, error) => {
      this.toast.show(error);
    });
    this.tabledata.addEventListener(['provider:added', 'provider:changed'], (sender, data) => {
      console.log(`data added: ${data}`);
    });
  }

  main() {
    super.main();
    // Request the static & database endpoints
    this.static.request('/', null, API_STATIC_DELTA);
    this.sqlite.request('/', null, API_SQLITE_DELTA);
  }
}
