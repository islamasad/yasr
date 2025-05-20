import 'vite/modulepreload-polyfill';
import Alpine from 'alpinejs';
import 'preline';

import './style.css'; // Tailwind styles

window.Alpine = Alpine;
Alpine.start();
