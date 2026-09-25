export interface DomainValidationResult {
  valid: boolean;
  normalized: string;
  host: string;
  error?: string;
  suggestion?: string;
}

const VALID_LABEL_RE = /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/;
const VALID_TLD_RE = /^[a-z]{2,24}$/;

/**
 * Validates a domain or URL string before dispatching any backend request.
 * Catches typos such as "xyz,com", spaces, missing TLDs, or invalid characters.
 */
export function validateDomainOrUrl(raw: string): DomainValidationResult {
  const trimmed = raw.trim();
  if (!trimmed) {
    return {
      valid: false,
      normalized: '',
      host: '',
      error: 'Please enter a valid domain name (for example, svelte.dev or cloudflare.com).',
    };
  }

  // Catch comma typo like "xyz,com"
  if (trimmed.includes(',')) {
    const suggestion = trimmed.replace(/,/g, '.').replace(/\s+/g, '');
    return {
      valid: false,
      normalized: '',
      host: '',
      error: `Invalid domain "${trimmed}" — contains a comma (,) instead of a dot (.).`,
      suggestion,
    };
  }

  if (/\s/.test(trimmed)) {
    const suggestion = trimmed.replace(/\s+/g, '');
    return {
      valid: false,
      normalized: '',
      host: '',
      error: `Invalid domain "${trimmed}" — domain names cannot contain spaces.`,
      suggestion,
    };
  }

  let clean = trimmed.toLowerCase();
  clean = clean.replace(/^https?:\/\//, '').replace(/^https?\/\//, '');

  const slashIdx = clean.indexOf('/');
  let host = slashIdx !== -1 ? clean.slice(0, slashIdx) : clean;
  const pathPart = slashIdx !== -1 ? clean.slice(slashIdx) : '';

  const colonIdx = host.indexOf(':');
  if (colonIdx !== -1) {
    host = host.slice(0, colonIdx);
  }

  if (!host.includes('.')) {
    return {
      valid: false,
      normalized: '',
      host: '',
      error: `"${host}" is missing a Top-Level Domain extension (such as .com, .ai, .dev, .io).`,
      suggestion: `${host}.com`,
    };
  }

  if (host.includes('..')) {
    const suggestion = host.replace(/\.+/g, '.') + pathPart;
    return {
      valid: false,
      normalized: '',
      host: '',
      error: `Invalid domain "${host}" — contains consecutive dots.`,
      suggestion,
    };
  }

  const labels = host.split('.');
  const tld = labels[labels.length - 1] || '';

  if (!VALID_TLD_RE.test(tld)) {
    return {
      valid: false,
      normalized: '',
      host: '',
      error: `Invalid TLD extension ".${tld}" in "${host}". Extensions must be 2–24 letters.`,
    };
  }

  for (const label of labels) {
    if (!VALID_LABEL_RE.test(label)) {
      return {
        valid: false,
        normalized: '',
        host: '',
        error: `Invalid hostname segment "${label}" in "${host}". Only letters, numbers, and interior hyphens are permitted.`,
      };
    }
  }

  return {
    valid: true,
    normalized: host + pathPart,
    host,
  };
}
