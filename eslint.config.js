// ESLint flat config — console guard for harden-ts-xss-residual
// no-console is enforced for ts files. Intentionally kept console calls
// must be DEV-gated (`if (import.meta.env.DEV) console.*`) or explicitly
// allowlisted with `// eslint-disable-next-line no-console` on the line
// above. Prefer DEV-gating; use eslint-disable only for intentional prod
// diagnostics that must remain in the bundle.
export default [
  {
    files: ["ts/**/*.{ts,js}"],
    rules: {
      "no-console": ["error", { allow: [] }],
    },
  },
];
