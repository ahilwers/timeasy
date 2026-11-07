// Keycloakify strips target attributes from message HTML. Keep the registration
// form open when users read the linked timeasy terms and privacy policy.
document.addEventListener("click", (event) => {
  if (!(event.target instanceof Element)) return;
  const link = event.target.closest('label[for="termsAccepted"] a');
  if (!(link instanceof HTMLAnchorElement)) return;

  const url = new URL(link.href);
  if (
    url.origin !== "https://timeasy.org" ||
    !/^\/(de|en)\/(terms|privacy)$/.test(url.pathname)
  ) {
    return;
  }

  link.target = "_blank";
  link.rel = "noopener noreferrer";
});
