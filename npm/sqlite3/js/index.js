
// CSS
import '../css/index.css';

// Application Controllers
import App from './controller/app';

// Views
import './view/nav-view';
import './view/nav-item-view';

import './view/table-view';
import './view/table-head-view';
import './view/table-body-view';

// Import js-framework
const jsf = require('@djthorpe/js-framework');

// Run
window.addEventListener('DOMContentLoaded', () => {
  const app = jsf.Controller.New(App);

  // Run the main function for the app
  console.log('Running application', app.constructor.name);
  app.main();
});
