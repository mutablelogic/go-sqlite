
// CSS
import '../css/index.css';

// Application Controllers
import App from './controller/app';

// Views
import ComponentView from './view/component-view';
import BadgeView from './view/badge-view';
import TableView from './view/table-view';
import TableHeadView from './view/table-head-view';
import TableBodyView from './view/table-body-view';

// Define tag names
ComponentView.define('badge-view', BadgeView);
ComponentView.define('table-view', TableView);
ComponentView.define('table-head-view', TableHeadView);
ComponentView.define('table-body-view', TableBodyView);

// Import js-framework
const jsf = require('@djthorpe/js-framework');

// Run
window.addEventListener('DOMContentLoaded', () => {
  const app = jsf.Controller.New(App);

  // Run the main function for the app
  console.log('Running application', app.constructor.name);
  app.main();
});

