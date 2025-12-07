import * as Turbo from "@hotwired/turbo"
import { Application } from "@hotwired/stimulus"
import TabsController from "./controllers/tabs_controller"

// Expose Turbo globally so Stimulus controllers can target frames.
window.Turbo = Turbo

const application = Application.start()
application.register("tabs", TabsController)

document.addEventListener("turbo:load", () => {
  console.log("Driplnk Frontend Loaded")
})
