import FingerprintJS from '@fingerprintjs/fingerprintjs'

// ... existing helper functions (setCookie, getCookie, uuidv4) remain ...

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

// Generate a UUID v4 as fallback
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

// Initialize visitor ID
export async function initAnalytics() {
    // 1. Check if we already have a visitor ID in cookie
    let visitorID = getCookie("drip_visitor");

    if (visitorID) {
        console.log("[Analytics] Existing Visitor ID:", visitorID);
        return;
    }

    try {
        // 2. Initialize FingerprintJS
        const fpPromise = FingerprintJS.load();
        const fp = await fpPromise;
        const result = await fp.get();

        // 3. Use the visitorId from FingerprintJS
        visitorID = result.visitorId;
        console.log("[Analytics] Generated Fingerprint ID:", visitorID);
    } catch (error) {
        console.error("[Analytics] Fingerprint failed:", error);
        // 4. Fallback to UUID if fingerprinting fails/blocked
        visitorID = uuidv4();
        console.log("[Analytics] Fallback to UUID:", visitorID);
    }

    // 5. Persist
    setCookie("drip_visitor", visitorID);
}
