/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["*.tmpl", "**/*.tmpl"],
  theme: {
    screens: {
      sm: "100%",
      md: `870px`,
    },
    fontFamily: {
      mono: ["Inconsolata", "monospace"],
    },
    extend: {
      colors: {
        // palette ported from patterns-of-attachment: dark teal surfaces,
        // orange accents and cyan highlights
        gray: {
          50: "rgb(6, 40, 40)",
          300: "rgb(255, 142, 83)",
          400: "rgb(255, 142, 83)",
          500: "#bbb",
          600: "rgb(240, 240, 240)",
          800: "rgb(240, 240, 240)",
          900: "rgb(255, 142, 83)",
        },
        white: "rgb(2, 26, 26)",
        blue: {
          500: "rgb(0, 183, 255)",
        },
        green: {
          200: "rgb(2, 60, 60)",
          500: "rgb(0, 255, 255)",
        },
        red: {
          200: "rgb(70, 30, 30)",
          400: "rgb(255, 107, 107)",
        },
      },
    },
  },
};
