/** @type {import('tailwindcss').Config} */
module.exports = {
  theme: {
    extend: {
      colors: {
        // Background and base text color
        solarLightBase03: "#f8f2e0",
        solarLightBase02: "#F2F1F9",
        solarLightBase01: "#93a1a1",
        solarLightBase00: "#839496",
        darkGray: "#333333",

        // Highlight color
        periwinkle: "#CCCCFF",
      },
      typography: {
        DEFAULT: {
          css: {
            h1: {
              fontWeight: "500",
            },
            h2: {
              fontWeight: "500",
            },
            h3: {
              fontWeight: "500",
            },
            h4: {
              fontWeight: "500",
            },
          },
        },
      },
    },
  },
  content: ["./templates/**/*.html", "./templates/*.html"],
  plugins: [require("@tailwindcss/typography")],
};
