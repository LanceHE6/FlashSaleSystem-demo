import { createApp } from 'vue';
import {router} from "./router/index";
import App from "@/App.vue";

import {createDiscreteApi} from "naive-ui";
import naive from "naive-ui";
import {createPinia} from "pinia";
import axios from "axios";

// 设置默认后端服务地址
axios.defaults.baseURL = "http://localhost:8080"

const {message, notification, dialog, loadingBar} = createDiscreteApi(["message", "notification", "dialog", "loadingBar"])

/*
 *axios
 *pinia
 * sass
 * vue-router
 * naive-ui
 * wangeditor
 */

const app = createApp(App)

app.provide("axios", axios)
app.provide("message", message)
app.provide("notification", notification)
app.provide("dialog", dialog)
app.provide("loadingBar", loadingBar)
app.provide("server_url", axios.defaults.baseURL)

app.use(naive)
app.use(router)
app.use(createPinia())

// // 配置全局拦截器，所有的请求都会经过这里添加token请求头
// const userStore = UserStore()
// axios.interceptors.request.use((config) =>{
//     config.headers.token = localStorage.getItem("token")
//     return config
// })

app.mount('#app')