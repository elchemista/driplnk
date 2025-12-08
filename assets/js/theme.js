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
    const html = document.documentElement;
    const currentTheme = html.getAttribute("data-theme");
    const pref = html.getAttribute("data-theme-preference");

    console.log("[Theme] Initializing theme", { currentTheme, pref });

    // If theme is explicitly set (light/dark), use it directly - already applied by data-theme
    if (currentTheme && currentTheme !== "" && currentTheme !== "system") {
        console.log(`[Theme] Using explicit theme: ${currentTheme}`);
        // The data-theme attribute is already set correctly in the HTML
        return;
    }

    // Only apply system theme if preference is "system"
    if (pref === "system" || currentTheme === "system" || currentTheme === "") {
        console.log("[Theme] Applying system theme");
        applySystemTheme();
        // Listen for OS changes
        const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
        mediaQuery.removeEventListener('change', applySystemTheme);
        mediaQuery.addEventListener('change', applySystemTheme);
    }
}

function applySystemTheme() {
    const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    const theme = isDark ? 'dark' : 'light';
    console.log(`[Theme] System preference: ${theme}`);
    document.documentElement.setAttribute('data-theme', theme);
}
