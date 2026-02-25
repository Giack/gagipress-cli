// Entry point — imports router which registers hashchange and load event listeners
// NOTE: All imports must use .js extensions (browser ES modules require explicit extensions)
import { dispatch } from './router.js';

// dispatch() is called automatically by the 'load' event listener in router.js
// No additional setup needed here — this file exists as the module entry point
// referenced by <script type="module" src="js/app.js"> in index.html
