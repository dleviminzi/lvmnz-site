/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: "class",
  theme: {
    extend: {
      fontFamily: {
        sans: ['"Geist"', "system-ui", "sans-serif"],
        mono: ['"Geist Mono"', "ui-monospace", "monospace"],
      },
      colors: {
        bg: "var(--bg)",
        text: "var(--text)",
        muted: "var(--muted)",
        border: "var(--border)",
        link: "var(--link)",
        "link-hover": "var(--link-hover)",
      },
      typography: {
        DEFAULT: {
          css: {
            "--tw-prose-body": "var(--text)",
            "--tw-prose-headings": "var(--text)",
            "--tw-prose-links": "var(--link)",
            "--tw-prose-bold": "var(--text)",
            "--tw-prose-counters": "var(--muted)",
            "--tw-prose-bullets": "var(--muted)",
            "--tw-prose-hr": "var(--border)",
            "--tw-prose-quotes": "var(--muted)",
            "--tw-prose-code": "var(--text)",
            "--tw-prose-pre-bg": "var(--code-bg)",
            "--tw-prose-pre-code": "var(--text)",
            "--tw-prose-th-borders": "var(--border)",
            "--tw-prose-td-borders": "var(--border)",
            h1: { fontWeight: "600" },
            h2: { fontWeight: "600" },
            h3: { fontWeight: "500" },
            h4: { fontWeight: "500" },
            a: {
              textDecoration: "none",
              "&:hover": { textDecoration: "underline" },
            },
            code: {
              fontFamily: '"Geist Mono", ui-monospace, monospace',
              fontSize: "0.875em",
            },
          },
        },
      },
      maxWidth: {
        content: "640px",
      },
    },
  },
  content: ["./templates/**/*.html", "./templates/*.html"],
  plugins: [require("@tailwindcss/typography")],
};
