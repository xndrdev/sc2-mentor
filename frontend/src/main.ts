import { createApp } from 'vue'
import { createPinia } from 'pinia'
import VueApexCharts from 'vue3-apexcharts'
import App from './App.vue'
import router from './router'
import './style.css'

function initUmamiTracking(): void {
  const scriptUrl = import.meta.env.VITE_UMAMI_SCRIPT_URL?.trim()
  const websiteId = import.meta.env.VITE_UMAMI_WEBSITE_ID?.trim()

  if (!scriptUrl || !websiteId) {
    return
  }

  const existingScript = document.querySelector<HTMLScriptElement>(
    `script[src="${scriptUrl}"][data-website-id="${websiteId}"]`,
  )

  if (existingScript) {
    return
  }

  const script = document.createElement('script')
  script.defer = true
  script.src = scriptUrl
  script.dataset.websiteId = websiteId
  document.head.appendChild(script)
}

initUmamiTracking()

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(VueApexCharts)

app.mount('#app')
