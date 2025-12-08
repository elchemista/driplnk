import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
  static values = {
    frameId: String,
    active: String
  }

  connect() {
    console.log("[TabsController] Controller connected", {
      frameId: this.frameIdValue,
      active: this.activeValue
    })

    if (!this.hasFrameIdValue) {
      this.frameIdValue = this.element.id || "dashboard-content"
    }

    if (this.activeValue) {
      this.updateHistory(this.activeValue)
    }
  }

  visit(event) {
    const url = event.currentTarget.getAttribute("href")
    const tab = event.params.tab || ""

    console.log(`🔗 [TabsController] Tab clicked:`, {
      tab,
      url,
      frameId: this.frameIdValue
    })

    if (tab) {
      this.activeValue = tab
      this.updateHistory(tab, url)
    }
  }

  updateHistory(tab, url) {
    const next = new URL(window.location.href)

    if (tab) {
      next.searchParams.set("tab", tab)
      next.hash = `#${tab}`
    } else {
      next.searchParams.delete("tab")
    }

    if (url) {
      const parsed = new URL(url, window.location.origin)
      if (parsed.hash) {
        next.hash = parsed.hash
      }
    }

    window.history.replaceState({}, "", next)
  }
}
