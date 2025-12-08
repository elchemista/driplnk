import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    connect() {
        // Auto-dismiss after 5 seconds
        this.timeout = setTimeout(() => {
            this.dismiss()
        }, 5000)
    }

    disconnect() {
        if (this.timeout) {
            clearTimeout(this.timeout)
        }
    }

    dismiss() {
        this.element.style.opacity = "0"
        this.element.style.transform = "translateY(-10px)"
        this.element.style.transition = "opacity 0.3s, transform 0.3s"
        setTimeout(() => {
            this.element.remove()
        }, 300)
    }
}
