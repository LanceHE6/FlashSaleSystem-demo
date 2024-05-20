import { createRouter, createWebHistory } from 'vue-router';

import FlashSale from "@/views/FlashSale.vue";



const routes = [
    { path: '/', component: FlashSale , name: 'home'},
    

];

const router = createRouter({
    history: createWebHistory(),
    routes,
});

export {router, routes};
