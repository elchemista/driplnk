import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    static values = {
        key: String
    }

    connect() {
        console.log("[FormAutosave] Controller connecting...", this.element)

        // Generate a unique key based on form action if not provided
        if (!this.hasKeyValue) {
            const formAction = this.element.action || this.element.getAttribute("action") || ""
            this.keyValue = `form_autosave_${formAction.replace(/[^a-zA-Z0-9]/g, "_")}`
        }

        console.log(`[FormAutosave] Using storage key: ${this.keyValue}`)

        // Restore saved data
        this.restore()

        // Set up auto-save on input changes
        this.element.addEventListener("input", this.handleInput.bind(this))
        console.log("[FormAutosave] Added input event listener")

        // Clear saved data on successful submission
        this.element.addEventListener("turbo:submit-end", this.handleSubmitEnd.bind(this))
        console.log("[FormAutosave] Added turbo:submit-end event listener")

        // Log Turbo Frame events to track navigation
        document.addEventListener("turbo:before-frame-render", this.logBeforeFrameRender.bind(this))
        document.addEventListener("turbo:frame-load", this.logFrameLoad.bind(this))
        console.log("[FormAutosave] Added Turbo Frame event listeners")
    }

    disconnect() {
        console.log(`[FormAutosave] Controller disconnecting for ${this.keyValue}`)
        this.element.removeEventListener("input", this.handleInput.bind(this))
        this.element.removeEventListener("turbo:submit-end", this.handleSubmitEnd.bind(this))
        document.removeEventListener("turbo:before-frame-render", this.logBeforeFrameRender.bind(this))
        document.removeEventListener("turbo:frame-load", this.logFrameLoad.bind(this))
    }

    logBeforeFrameRender(event) {
        console.log("🔄 [Turbo] Before frame render", event.target.id, event.detail)
    }

    logFrameLoad(event) {
        console.log("✅ [Turbo] Frame loaded", event.target.id)
    }

    handleInput(event) {
        console.log(`[FormAutosave] Input detected in field: ${event.target.name}`, event.target.value)

        // Debounce the save operation
        clearTimeout(this.saveTimeout)
        this.saveTimeout = setTimeout(() => {
            this.save()
        }, 300)
    }

    save() {
        const formData = new FormData(this.element)
        const data = {}

        for (let [key, value] of formData.entries()) {
            data[key] = value
        }

        console.log(`[FormAutosave] Saving form data:`, data)

        try {
            localStorage.setItem(this.keyValue, JSON.stringify(data))
            console.log(`✅ [FormAutosave] Successfully saved to localStorage: ${this.keyValue}`)
            console.log(`[FormAutosave] Current localStorage:`, localStorage.getItem(this.keyValue))
        } catch (e) {
            console.error("❌ [FormAutosave] Failed to save to localStorage:", e)
        }
    }

    restore() {
        console.log(`[FormAutosave] Attempting to restore data from: ${this.keyValue}`)

        try {
            const savedData = localStorage.getItem(this.keyValue)

            if (!savedData) {
                console.log(`[FormAutosave] No saved data found for ${this.keyValue}`)
                return
            }

            console.log(`[FormAutosave] Found saved data:`, savedData)
            const data = JSON.parse(savedData)
            console.log(`[FormAutosave] Parsed data:`, data)

            let restoredCount = 0

            // Restore each field
            for (let [name, value] of Object.entries(data)) {
                const field = this.element.querySelector(`[name="${name}"]`)

                if (field) {
                    if (field.type === "checkbox") {
                        field.checked = value === "on" || value === "true"
                        console.log(`[FormAutosave] Restored checkbox ${name}: ${field.checked}`)
                    } else if (field.type === "radio") {
                        const radioButton = this.element.querySelector(`[name="${name}"][value="${value}"]`)
                        if (radioButton) {
                            radioButton.checked = true
                            console.log(`[FormAutosave] Restored radio ${name}: ${value}`)
                        }
                    } else {
                        field.value = value
                        console.log(`[FormAutosave] Restored field ${name}: ${value}`)
                    }
                    restoredCount++
                } else {
                    console.warn(`[FormAutosave] Field not found: ${name}`)
                }
            }

            console.log(`✅ [FormAutosave] Restored ${restoredCount} fields from ${this.keyValue}`)
        } catch (e) {
            console.error("❌ [FormAutosave] Failed to restore from localStorage:", e)
        }
    }

    handleSubmitEnd(event) {
        console.log(`[FormAutosave] Form submitted. Success: ${event.detail.success}`)

        // Only clear on successful submission
        if (event.detail.success) {
            this.clear()
        }
    }

    clear() {
        try {
            localStorage.removeItem(this.keyValue)
            console.log(`✅ [FormAutosave] Cleared saved data for ${this.keyValue}`)
        } catch (e) {
            console.error("❌ [FormAutosave] Failed to clear localStorage:", e)
        }
    }
}
