/**
 * Hash-based SPA router.
 *
 * Syncs location.hash with the current view so users get bookmarks,
 * back/forward navigation, and shareable deep links.
 *
 * Route format: #/view[/param]
 * Examples:     #/characters, #/sheet/42, #/compendium
 *
 * Integrates with the existing showView()/navigation.ts system:
 *  - showView() → calls Router.navigate() → updates hash
 *  - hashchange → calls showView() → shows the view
 *
 * Login redirect: if the user is not authenticated, the server
 * serves the login page regardless of hash. After login, the SPA
 * reads the hash and navigates to the intended view.
 */

import type { ViewState } from './types';

export interface Route {
  view: ViewState;
  params: Record<string, string>;
}

let currentRoute: Route = { view: 'characters', params: {} };
let routerInitialized = false;

/**
 * Parse the current location.hash into a Route.
 */
export function parseHash(hash: string): Route {
  const h = hash.replace(/^#\/?/, '');
  if (!h) return { view: 'characters', params: {} };

  const parts = h.split('/');
  const view = parts[0] as ViewState;
  const params: Record<string, string> = {};

  // View-specific param extraction
  if ((view === 'sheet' || view === 'singleEncounter') && parts[1]) {
    params.id = parts[1];
  }

  return { view, params };
}

/**
 * Serialize a Route to a hash string.
 */
export function routeToHash(route: Route): string {
  const { view, params } = route;
  if (params.id) return `#/${view}/${params.id}`;
  return `#/${view}`;
}

/**
 * Navigate to a view, updating the hash and returning the new Route.
 */
export function navigate(view: ViewState, params: Record<string, string> = {}): Route {
  const route: Route = { view, params };
  currentRoute = route;
  const hash = routeToHash(route);
  if (location.hash !== hash) {
    location.hash = hash;
  }
  return route;
}

/**
 * Get the current Route.
 */
export function getCurrentRoute(): Route {
  return currentRoute;
}

/**
 * Initialize the router: set up hashchange listener and process the
 * initial hash. Call once after the app's init() has completed.
 */
let routerHandler: ((e: HashChangeEvent) => void) | null = null;

export function initRouter(onNavigate: (route: Route) => void): void {
  if (routerInitialized) return;
  routerInitialized = true;

  // Handle hash changes (back/forward, bookmarks)
  routerHandler = () => {
    const route = parseHash(location.hash);
    currentRoute = route;
    onNavigate(route);
  };
  window.addEventListener('hashchange', routerHandler);
}

/**
 * Reset router state (for testing only).
 */
export function resetRouter(): void {
  routerInitialized = false;
  currentRoute = { view: 'characters', params: {} };
  if (routerHandler) {
    window.removeEventListener('hashchange', routerHandler);
    routerHandler = null;
  }
}

/**
 * Read the initial hash and navigate to it. Returns the parsed Route
 * or null if no hash was set.
 */
export function navigateToInitialHash(onNavigate: (route: Route) => void): Route | null {
  if (location.hash) {
    const route = parseHash(location.hash);
    currentRoute = route;
    onNavigate(route);
    return route;
  }
  return null;
}
