import * as Turbo from "@hotwired/turbo"
import { Application } from "@hotwired/stimulus"
import TabsController from "./controllers/tabs_controller"
import FlashController from "./controllers/flash_controller"
import FormAutosaveController from "./controllers/form_autosave_controller"
import { initAnalytics } from "./analytics.js"

// Expose Turbo globally so Stimulus controllers can target frames.
window.Turbo = Turbo

// Initialize Analytics (Visitor ID)
initAnalytics()

const application = Application.start()
application.register("tabs", TabsController)
application.register("flash", FlashController)
application.register("form-autosave", FormAutosaveController)

// Helper to show flash messages
function showFlash(message, type = "info") {
  const container = document.getElementById("flash-messages")
  if (!container) return

  const alertDiv = document.createElement("div")
  alertDiv.dataset.controller = "flash"
  alertDiv.className = `alert alert-${type} shadow-lg mb-2`

  // Icon based on type
  let icon = ""
  switch (type) {
    case "success":
      icon = `<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`
      break
    case "error":
      icon = `<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`
      break
    case "warning":
      icon = `<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>`
      break
    default:
      icon = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-info shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>`
  }

  alertDiv.innerHTML = `
    <div>
      ${icon}
      <span>${message}</span>
    </div>
  `

  container.appendChild(alertDiv)
}

document.addEventListener("turbo:load", () => {
  console.log("Driplnk Frontend Loaded")
})

// Global Network Handlers
window.addEventListener("offline", () => {
  showFlash("Connection lost. You are now offline.", "warning")
})

window.addEventListener("online", () => {
  showFlash("Connection restored. You are back online.", "success")
})

document.addEventListener("turbo:fetch-request-error", (event) => {
  console.error("Turbo fetch error:", event)
  showFlash("Network error. Please check your connection.", "error")
  // Prevent Turbo from throwing unhandled rejection if possible, 
  // though Turbo usually logs it anyway.
})

document.addEventListener("turbo:before-fetch-request", (event) => {
  const token = document.querySelector('meta[name="csrf-token"]')?.getAttribute("content")
  if (token) {
    event.detail.fetchOptions.headers["X-CSRF-Token"] = token
  }
})

