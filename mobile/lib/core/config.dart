/// App-wide configuration that is not user-specific.
library;

/// The hosted Villum instance offered on first run.
///
/// The onboarding screen prefills this address but leaves it editable so users
/// can point the app at their own self-hosted DnD server. A previously saved
/// server address always takes precedence over this default.
const String defaultServerUrl = 'https://dnd.vandijke.xyz';
