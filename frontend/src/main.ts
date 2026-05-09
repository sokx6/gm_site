import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './style.css'
import './assets/styles/main.css'
import App from './App.vue'
import router from './router'
import { useAuthStore } from '@/stores/auth'

const app = createApp(App)

app.use(createPinia())
app.use(router)

// Restore session from token before mounting
const auth = useAuthStore()
auth.tryRestoreSession().then(() => {
  app.mount('#app')
})
