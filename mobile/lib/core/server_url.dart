/// Server-address helpers.
///
/// Villum has no hosted service and no default address: the user always types
/// the server they want to talk to, so the app never guesses one.
library;

/// Strips surrounding whitespace and any trailing slashes.
///
/// The address is concatenated with absolute API paths (`/api/...`) in a few
/// places, so `https://host/` and `https://host` must not produce `//api/...`.
String normalizeServerUrl(String raw) {
  var url = raw.trim();
  final schemeEnd = url.indexOf('://');
  while (url.length > 1 &&
      url.endsWith('/') &&
      (schemeEnd < 0 || url.length > schemeEnd + 3)) {
    url = url.substring(0, url.length - 1);
  }
  return url;
}

/// Returns a user-facing message when [raw] is not a usable http(s) address,
/// or `null` when it is.
String? serverUrlError(String raw) {
  final url = normalizeServerUrl(raw);
  if (url.isEmpty) return 'Enter your Villum server address.';
  final uri = Uri.tryParse(url);
  if (uri == null ||
      (uri.scheme != 'http' && uri.scheme != 'https') ||
      uri.host.isEmpty) {
    return 'Enter a valid URL starting with http:// or https://';
  }
  return null;
}
