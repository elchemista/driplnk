document.addEventListener("turbo:load", () => {
    initTheme();
});

// Also run on initial load
if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initTheme);
} else {
    initTheme();
}

function initTheme() {
    // Only intervene if system usage is indicated (e.g. by emptiness or explicit attribute)
    // We check for data-theme-preference="system"
    const html = document.documentElement;
    const pref = html.getAttribute("data-theme-preference");

    if (pref === "system") {
        applySystemTheme();
        // Listen for OS changes
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', applySystemTheme);
    } else {
        // If hardcoded light/dark, remove listener if exists? 
        // For simplicity, just leave it.
    }
}

function applySystemTheme() {
    const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    document.documentElement.setAttribute('data-theme', isDark ? 'dark' : 'light');
}
