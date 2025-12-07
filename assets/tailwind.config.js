module.exports = {
    content: ["../views/**/*.templ"],
    theme: {
        extend: {},
    },
    plugins: [require("daisyui")],
    daisyui: {
        themes: ["light", "dark"],
    },
}
