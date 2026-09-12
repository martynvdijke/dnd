declare module 'bootstrap';

// htmx is loaded globally by the page shell (see layout templates).
declare const htmx: {
  trigger(el: Element, name: string): void;
  ajax(verb: string, path: string, ctx?: Record<string, unknown>): void;
};
