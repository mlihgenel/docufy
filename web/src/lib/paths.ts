const stripTrailingSlash = (value: string): string =>
  value.length > 1 && value.endsWith("/") ? value.slice(0, -1) : value;

export const withBase = (href: string): string => {
  const baseUrl = stripTrailingSlash(import.meta.env.BASE_URL || "/");
  const path = href.startsWith("/") ? href : `/${href}`;
  if (baseUrl === "/") {
    return path;
  }
  return `${baseUrl}${path}`.replace(/\/{2,}/g, "/");
};

export const normalizePathname = (pathname: string): string => {
  const baseUrl = stripTrailingSlash(import.meta.env.BASE_URL || "/");
  let value = pathname || "/";

  if (!value.startsWith("/")) {
    value = `/${value}`;
  }
  if (!value.endsWith("/")) {
    value = `${value}/`;
  }

  if (baseUrl !== "/" && value.startsWith(`${baseUrl}/`)) {
    value = value.slice(baseUrl.length);
    if (!value.startsWith("/")) {
      value = `/${value}`;
    }
  }

  return value;
};
