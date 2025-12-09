/**
 * Analytics Helper
 * Manages visitor identity for server-side analytics.
 */

// Generate a UUID v4
function uuidv4() {
    return "10000000-1000-4000-8000-100000000000".replace(/[018]/g, c =>
        (+c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> +c / 4).toString(16)
    );
}

// Set a cookie (1 year expiry)
function setCookie(name, value) {
    const d = new Date();
    d.setTime(d.getTime() + (365 * 24 * 60 * 60 * 1000));
    let expires = "expires=" + d.toUTCString();
    document.cookie = name + "=" + value + ";" + expires + ";path=/;SameSite=Lax";
}

// Get a cookie by name
function getCookie(name) {
    let nameEQ = name + "=";
    let ca = document.cookie.split(';');
    for (let i = 0; i < ca.length; i++) {
        let c = ca[i];
        while (c.charAt(0) == ' ') c = c.substring(1, c.length);
        if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length, c.length);
    }
    return null;
}

// Initialize visitor ID
export function initAnalytics() {
    let visitorID = getCookie("drip_visitor");
    if (!visitorID) {
        visitorID = uuidv4();
        setCookie("drip_visitor", visitorID);
        console.log("[Analytics] New visitor ID assigned:", visitorID);
    } else {
        console.log("[Analytics] Visitor ID:", visitorID);
    }
}
