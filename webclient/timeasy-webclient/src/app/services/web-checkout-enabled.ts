/**
 * Build-time switch for the Paddle web checkout.
 *
 * Paddle currently blocks transaction creation for this seller
 * (`transaction_creation_blocked`), so every checkout attempt fails. While this
 * is `false`, the upgrade page offers the iOS app subscription instead of an
 * embedded checkout. Set it to `true` once Paddle re-enables checkout for the
 * seller and rebuild the client — no other code change is required.
 */
export const WEB_CHECKOUT_ENABLED = false;

/** App Store listing of the iOS app, localized for the German and English UI. */
export function appStoreUrl(lang: string | undefined): string {
  return lang === 'de'
    ? 'https://apps.apple.com/de/app/id1620757668'
    : 'https://apps.apple.com/app/id1620757668';
}
