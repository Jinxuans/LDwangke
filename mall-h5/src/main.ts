import { createApp } from 'vue';
import 'vant/lib/index.css';
import './style.css';
import App from './App.vue';
import router from './router';

createApp(App).use(router).mount('#app');
