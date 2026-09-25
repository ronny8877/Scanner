import type {
  ScanReport,
  DomainInquiry,
  CrawlReport,
  Valuation,
  ScanResultItem,
  HistoryReport,
  ReconReport,
  SavedDomain,
  Job,
  ParallelSuiteResult,
  RobotsSitemapReport,
  MetaSocialReport,
} from './types';

const API_BASE = 'http://localhost:8080';

let currentAbortController: AbortController | null = null;

function createSignal(): AbortSignal {
  if (currentAbortController) {
    currentAbortController.abort();
  }
  currentAbortController = new AbortController();
  return currentAbortController.signal;
}

export async function cancelRunningJob(jobId = 'all'): Promise<Job[]> {
  if (currentAbortController) {
    currentAbortController.abort();
    currentAbortController = null;
  }
  try {
    const res = await fetch(`${API_BASE}/api/jobs/cancel`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: jobId }),
    });
    if (!res.ok) return [];
    const data = await res.json();
    return data.jobs ?? [];
  } catch {
    return [];
  }
}

export async function checkBackendHealth(): Promise<boolean> {
  try {
    const res = await fetch(`${API_BASE}/api/health`, { signal: AbortSignal.timeout(2000) });
    return res.ok;
  } catch {
    return false;
  }
}

export async function runDomainScan(params: {
  keywords: string[];
  tlds: string[];
  dictionaryPack?: string;
  mutations: boolean;
  onlyAvailable: boolean;
  minScore: number;
}): Promise<{ report: ScanReport; liveBackend: boolean }> {
  const signal = createSignal();
  try {
    const res = await fetch(`${API_BASE}/api/scan`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      signal,
      body: JSON.stringify({
        keywords: params.keywords,
        tlds: params.tlds,
        dictionaryPack: params.dictionaryPack || '',
        mutations: params.mutations,
        onlyAvailable: params.onlyAvailable,
        minScore: params.minScore,
        concurrency: 18,
        maxResults: 64,
      }),
    });
    if (!res.ok) throw new Error('Backend scan failed');
    const report: ScanReport = await res.json();
    return { report, liveBackend: true };
  } catch {
    return { report: synthesizeScanReport(params), liveBackend: false };
  }
}

export async function runDomainInspect(domainInput: string): Promise<{ inquiry: DomainInquiry; liveBackend: boolean }> {
  const signal = createSignal();
  try {
    const res = await fetch(`${API_BASE}/api/inspect?domain=${encodeURIComponent(domainInput)}`, { signal });
    if (!res.ok) throw new Error('Backend inspect failed');
    const inquiry: DomainInquiry = await res.json();
    return { inquiry, liveBackend: true };
  } catch {
    return { inquiry: synthesizeDomainInquiry(domainInput), liveBackend: false };
  }
}

export async function runDomainHistory(domainInput: string): Promise<{ history: HistoryReport; liveBackend: boolean }> {
  const signal = createSignal();
  try {
    const res = await fetch(`${API_BASE}/api/history?domain=${encodeURIComponent(domainInput)}`, { signal });
    if (!res.ok) throw new Error('Backend history failed');
    const history: HistoryReport = await res.json();
    return { history, liveBackend: true };
  } catch {
    return { history: synthesizeHistoryReport(domainInput), liveBackend: false };
  }
}

export async function runPortRecon(domainInput: string): Promise<{ recon: ReconReport; liveBackend: boolean }> {
  const signal = createSignal();
  try {
    const res = await fetch(`${API_BASE}/api/recon?domain=${encodeURIComponent(domainInput)}`, { signal });
    if (!res.ok) throw new Error('Backend recon failed');
    const recon: ReconReport = await res.json();
    return { recon, liveBackend: true };
  } catch {
    return { recon: synthesizeReconReport(domainInput), liveBackend: false };
  }
}

export async function runSiteCrawl(params: {
  targetUrl: string;
  maxPages: number;
  maxDepth: number;
}): Promise<{ report: CrawlReport; liveBackend: boolean }> {
  const signal = createSignal();
  try {
    const res = await fetch(`${API_BASE}/api/crawl`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      signal,
      body: JSON.stringify(params),
    });
    if (!res.ok) throw new Error('Backend crawl failed');
    const report: CrawlReport = await res.json();
    return { report, liveBackend: true };
  } catch {
    return { report: synthesizeCrawlReport(params.targetUrl), liveBackend: false };
  }
}

export async function runRobotsSitemapCheck(
  targetInput: string
): Promise<{ report: RobotsSitemapReport; liveBackend: boolean }> {
  const signal = createSignal();
  try {
    const res = await fetch(`${API_BASE}/api/robots?target=${encodeURIComponent(targetInput)}`, { signal });
    if (!res.ok) throw new Error('Robots/Sitemap check failed');
    const report: RobotsSitemapReport = await res.json();
    return { report, liveBackend: true };
  } catch {
    return { report: synthesizeRobotsSitemapReport(targetInput), liveBackend: false };
  }
}

export async function runMetaSocialCheck(
  targetInput: string
): Promise<{ report: MetaSocialReport; liveBackend: boolean }> {
  const signal = createSignal();
  try {
    const res = await fetch(`${API_BASE}/api/meta?target=${encodeURIComponent(targetInput)}`, { signal });
    if (!res.ok) throw new Error('Meta/Social check failed');
    const report: MetaSocialReport = await res.json();
    return { report, liveBackend: true };
  } catch {
    return { report: synthesizeMetaSocialReport(targetInput), liveBackend: false };
  }
}

export async function runFullParallelSuite(
  domainInput: string
): Promise<{ suite: ParallelSuiteResult; liveBackend: boolean }> {
  const signal = createSignal();
  try {
    const res = await fetch(`${API_BASE}/api/parallel-suite`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      signal,
      body: JSON.stringify({ domain: domainInput }),
    });
    if (!res.ok) throw new Error('Parallel suite failed');
    const suite: ParallelSuiteResult = await res.json();
    return { suite, liveBackend: true };
  } catch {
    const clean = domainInput.trim().toLowerCase();
    return {
      suite: {
        domain: clean,
        inquiry: synthesizeDomainInquiry(clean),
        history: synthesizeHistoryReport(clean),
        recon: synthesizeReconReport(clean),
        crawl: synthesizeCrawlReport(clean),
      },
      liveBackend: false,
    };
  }
}

export async function fetchJobQueue(): Promise<Job[]> {
  try {
    const res = await fetch(`${API_BASE}/api/jobs`);
    if (!res.ok) return [];
    const data = await res.json();
    return data.jobs ?? [];
  } catch {
    return [];
  }
}

export async function fetchWatchlist(): Promise<SavedDomain[]> {
  try {
    const res = await fetch(`${API_BASE}/api/watchlist`);
    if (!res.ok) throw new Error('Watchlist fetch failed');
    const data = await res.json();
    return data.items ?? [];
  } catch {
    const raw = localStorage.getItem('scanner_watchlist');
    if (raw) {
      try {
        return JSON.parse(raw);
      } catch {
        // ignore
      }
    }
    return [];
  }
}

export async function saveDomainToVault(params: {
  domain: string;
  available?: boolean;
  notes?: string;
  tags?: string[];
}): Promise<SavedDomain[]> {
  try {
    const res = await fetch(`${API_BASE}/api/watchlist`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(params),
    });
    if (!res.ok) throw new Error('Failed to save');
    return await fetchWatchlist();
  } catch {
    const current = await fetchWatchlist();
    const now = new Date().toISOString();
    const clean = params.domain.toLowerCase().trim();
    const filtered = current.filter((i) => i.domain !== clean);
    const next: SavedDomain[] = [
      {
        domain: clean,
        available: params.available ?? true,
        status: params.available ? 'Available' : 'Registered',
        valuation: evaluateLocal(clean, params.available ?? true),
        notes: params.notes || 'Saved in studio session',
        tags: params.tags?.length ? params.tags : ['Shortlist'],
        savedAt: now,
        lastCheckedAt: now,
      },
      ...filtered,
    ];
    localStorage.setItem('scanner_watchlist', JSON.stringify(next));
    return next;
  }
}

export async function removeDomainFromVault(domainStr: string): Promise<SavedDomain[]> {
  try {
    const res = await fetch(`${API_BASE}/api/watchlist?domain=${encodeURIComponent(domainStr)}`, {
      method: 'DELETE',
    });
    if (!res.ok) throw new Error('Delete failed');
    const data = await res.json();
    return data.items ?? [];
  } catch {
    const current = await fetchWatchlist();
    const next = current.filter((i) => i.domain !== domainStr.toLowerCase().trim());
    localStorage.setItem('scanner_watchlist', JSON.stringify(next));
    return next;
  }
}

export async function recheckWatchlistParallel(): Promise<SavedDomain[]> {
  try {
    const res = await fetch(`${API_BASE}/api/watchlist/recheck`, { method: 'POST' });
    if (!res.ok) throw new Error('Recheck failed');
    const data = await res.json();
    return data.items ?? [];
  } catch {
    return await fetchWatchlist();
  }
}

// --- Client-side fallback synthesis with calibrated pricing ---

const tldRegFees: Record<string, number> = {
  com: 10, ai: 68, io: 38, dev: 12, co: 24, app: 14, net: 12, org: 11,
  sh: 32, gg: 42, so: 48, cloud: 16, tech: 14, studio: 22, design: 28,
  tools: 24, codes: 28, xyz: 2, me: 16, vc: 55, finance: 35, store: 8, shop: 8,
};

function evaluateLocal(domainStr: string, isAvailable = true): Valuation {
  const [name = 'nova', tld = 'com'] = domainStr.toLowerCase().split('.');
  const regFee = tldRegFees[tld] ?? 12;
  const lenScore = name.length <= 4 ? 27 : name.length <= 6 ? 23 : name.length <= 8 ? 18 : 12;
  const tldMap: Record<string, number> = { com: 25, ai: 24, io: 21, dev: 19, co: 19, app: 18 };
  const tldScore = tldMap[tld] ?? 14;
  const phoneticScore = 21;
  const keywordScore = 15;
  const score = Math.min(99, lenScore + tldScore + phoneticScore + keywordScore);

  let tier = 'High-Potential Brandable';
  let minUsd = 95;
  let maxUsd = 340;
  let estimatedDisplay = `$${regFee}/yr · Flip $${minUsd}–$${maxUsd}`;

  if (isAvailable) {
    if (score >= 84) {
      tier = 'Prime Unclaimed Gem';
      minUsd = 350;
      maxUsd = 1250;
      estimatedDisplay = `$${regFee}/yr · Flip $350–$1,250`;
    }
  } else {
    if (name.length <= 5 && tld === 'com') {
      tier = 'Institutional .COM Asset';
      minUsd = 35000;
      maxUsd = 140000;
      estimatedDisplay = '$35,000 - $140,000+';
    } else {
      tier = 'Established Brand Domain';
      minUsd = 1200;
      maxUsd = 4200;
      estimatedDisplay = '$1,200 - $4,200';
    }
  }

  return {
    score,
    tier,
    regFeeUsd: regFee,
    regFeeDisplay: `$${regFee}/yr Reg`,
    estimatedMinUsd: minUsd,
    estimatedMaxUsd: maxUsd,
    estimatedDisplay,
    lengthScore: lenScore,
    tldScore,
    phoneticScore,
    keywordScore,
    highlights: [
      name.length <= 7 ? `Compact ${name.length}-letter root` : 'Clean brandable cadence',
      tld === 'com' ? `Global .com standard ($${regFee}/yr reg)` : `.${tld} tech extension ($${regFee}/yr reg)`,
      'Natural vowel-consonant phonetic cadence',
    ],
  };
}

function synthesizeScanReport(params: {
  keywords: string[];
  tlds: string[];
  dictionaryPack?: string;
  mutations: boolean;
  onlyAvailable: boolean;
  minScore: number;
}): ScanReport {
  const seeds = params.keywords.length ? params.keywords : ['veltrix', 'nova'];
  const suffixes = params.mutations ? ['', 'hq', 'labs', 'flow', 'grid', 'core', 'studio'] : [''];
  const items: ScanResultItem[] = [];

  for (const seed of seeds) {
    const clean = seed.toLowerCase().replace(/[^a-z0-9]/g, '');
    for (const sfx of suffixes) {
      const root = `${clean}${sfx}`;
      for (const tld of params.tlds) {
        const full = `${root}.${tld}`;
        const available = sfx !== '' || tld === 'ai' || tld === 'dev' || root.length >= 7;
        const val = evaluateLocal(full, available);
        if (params.onlyAvailable && !available) continue;
        if (params.minScore > 0 && val.score < params.minScore) continue;

        items.push({
          domain: full,
          rootName: root,
          tld,
          available,
          status: available ? 'Available' : 'Registered',
          nameservers: available ? [] : ['ns1.cloudflare.com', 'ns2.cloudflare.com'],
          valuation: val,
          latencyMs: 18 + ((full.length * 7) % 45),
        });
      }
    }
  }

  items.sort((a, b) => {
    if (a.available !== b.available) return a.available ? -1 : 1;
    return b.valuation.score - a.valuation.score;
  });

  const availableCount = items.filter((i) => i.available).length;
  return {
    seedKeywords: seeds,
    dictionaryUsed: params.dictionaryPack,
    totalChecked: items.length,
    availableCount,
    takenCount: items.length - availableCount,
    highValueCount: items.filter((i) => i.available && i.valuation.score >= 73).length,
    durationMs: 240,
    items,
  };
}

function synthesizeDomainInquiry(raw: string): DomainInquiry {
  const clean = raw.toLowerCase().replace(/^https?:\/\//, '').split('/')[0] || 'svelte.dev';
  const full = clean.includes('.') ? clean : `${clean}.com`;
  const isAvailable = full.includes('hq') || full.includes('veltrix');
  return {
    domain: full,
    available: isAvailable,
    statusSummary: isAvailable ? 'Available' : 'Registered',
    liveSiteUrl: `https://${full}`,
    waybackCalendarUrl: `https://web.archive.org/web/*/${full}`,
    registeredAt: isAvailable ? undefined : '2008-08-15T14:22:10Z',
    updatedAt: isAvailable ? undefined : '2025-11-02T09:15:00Z',
    expiresAt: isAvailable ? undefined : '2028-08-15T14:22:10Z',
    domainAge: isAvailable ? undefined : '18y 1m (6,615 days)',
    daysToExpiry: isAvailable ? undefined : 689,
    registrar: isAvailable ? undefined : 'Cloudflare, Inc. (IANA #1910)',
    registryHandle: isAvailable ? undefined : 'DOM-8849201-VRSN',
    dnssec: isAvailable ? undefined : 'Signed (DNSSEC Active)',
    statusFlags: isAvailable ? [] : ['clientTransferProhibited', 'clientUpdateProhibited'],
    nameservers: isAvailable ? [] : ['ada.ns.cloudflare.com', 'bob.ns.cloudflare.com'],
    dns: isAvailable
      ? {}
      : {
          a: ['104.21.44.19', '172.67.188.91'],
          aaaa: ['2606:4700:3031::ac43:bc5b'],
          cname: `www.${full}`,
          ptr: ['edge-node.cloudflare.com'],
          mx: ['10 mx1.forwardemail.net', '20 mx2.forwardemail.net'],
          ns: ['ada.ns.cloudflare.com', 'bob.ns.cloudflare.com'],
          txt: ['v=spf1 include:_spf.google.com ~all', 'google-site-verification=Tk92YV8xOTk'],
          dmarc: ['v=DMARC1; p=quarantine; rua=mailto:dmarc@' + full],
          spf: 'v=spf1 include:_spf.google.com ~all',
        },
    valuation: evaluateLocal(full, isAvailable),
    checkedAt: new Date().toISOString(),
    checkLatencyMs: 64,
  };
}

function synthesizeHistoryReport(raw: string): HistoryReport {
  const clean = raw.toLowerCase().replace(/^https?:\/\//, '').split('/')[0] || 'svelte.dev';
  const isVirgin = clean.includes('veltrix') || clean.includes('nexora');
  if (isVirgin) {
    return {
      domain: clean,
      currentlyRegistered: false,
      previouslyRegistered: false,
      historyVerdict: 'Clean Virgin Domain (Never Registered in Past)',
      summaryNote: `${clean} is completely unclaimed with zero prior registration records in RDAP, zero Wayback Machine snapshots, and zero historical SSL certificates.`,
      liveSiteUrl: `https://${clean}`,
      waybackCalendarUrl: `https://web.archive.org/web/*/${clean}`,
      totalSpanYears: 0,
      waybackSnapshots: 0,
      certCount: 0,
      checkLatencyMs: 112,
    };
  }
  const years = ['2008', '2010', '2012', '2014', '2016', '2018', '2020', '2022', '2024', '2026'];
  return {
    domain: clean,
    currentlyRegistered: true,
    previouslyRegistered: true,
    historyVerdict: 'Active & Historically Established Since 2008 (19 Years)',
    summaryNote: `${clean} is actively registered (created 2008-08-15) with 19 years of web history (2008–2026) and archived snapshots on Wayback Machine.`,
    liveSiteUrl: `https://${clean}`,
    waybackCalendarUrl: `https://web.archive.org/web/*/${clean}`,
    rdapCreatedDate: '2008-08-15',
    firstSeenAt: '2008-08-15',
    lastSeenAt: '2026-09-25',
    firstSeenYear: 2008,
    lastSeenYear: 2026,
    totalSpanYears: 19,
    waybackSnapshots: 228,
    activeYears: years,
    snapshots: years
      .slice()
      .reverse()
      .map((yr) => ({
        year: yr,
        date: `${yr}-06-15`,
        timestamp: `${yr}0615000000`,
        archiveUrl: `https://web.archive.org/web/${yr}0615000000/https://${clean}`,
        statusCode: '200',
      })),
    certCount: 24,
    pastIssuers: ["Let's Encrypt", 'Cloudflare Inc', 'Google Trust Services'],
    pastSubdomains: [`www.${clean}`, `cdn.${clean}`, `static.${clean}`],
    milestones: [
      { date: '2008-08-15', source: 'RDAP Registry', event: 'Domain registered in authoritative registry' },
      {
        date: '2008-10-12',
        source: 'Wayback Archive',
        event: 'Earliest archived web capture on 2008-10-12',
        archiveUrl: `https://web.archive.org/web/20081012000000/https://${clean}`,
      },
      {
        date: '2026-09-01',
        source: 'Wayback Archive',
        event: 'Latest Wayback snapshot captured (19 active archive years)',
        archiveUrl: `https://web.archive.org/web/20260901000000/https://${clean}`,
      },
    ],
    checkLatencyMs: 184,
  };
}

function synthesizeReconReport(raw: string): ReconReport {
  const clean = raw.toLowerCase().replace(/^https?:\/\//, '').split('/')[0] || 'svelte.dev';
  return {
    domain: clean,
    liveSiteUrl: `https://${clean}`,
    waybackCalendarUrl: `https://web.archive.org/web/*/${clean}`,
    targetIp: '104.21.44.19',
    reversePtr: 'edge-proxy.cloudflare.com',
    openPortsCount: 3,
    portsScanned: 14,
    ports: [
      { port: 80, service: 'HTTP', protocol: 'TCP', open: true, latencyMs: 19, category: 'Web', riskNote: 'Standard web traffic' },
      { port: 443, service: 'HTTPS', protocol: 'TCP', open: true, latencyMs: 18, category: 'Web', riskNote: 'Encrypted TLS web traffic' },
      { port: 8443, service: 'HTTPS Alt', protocol: 'TCP', open: true, latencyMs: 24, category: 'Web', riskNote: 'Alternative TLS web service' },
      { port: 22, service: 'SSH', protocol: 'TCP', open: false, latencyMs: 120, category: 'Remote Access' },
      { port: 3306, service: 'MySQL', protocol: 'TCP', open: false, latencyMs: 120, category: 'Database' },
      { port: 5432, service: 'PostgreSQL', protocol: 'TCP', open: false, latencyMs: 120, category: 'Database' },
      { port: 6379, service: 'Redis', protocol: 'TCP', open: false, latencyMs: 120, category: 'Database' },
    ],
    tls: {
      supported: true,
      version: 'TLS 1.3',
      cipherSuite: 'TLS_AES_128_GCM_SHA256',
      issuer: 'Google Trust Services WE1',
      subject: clean,
      validFrom: '2026-07-01',
      validUntil: '2026-10-01',
      daysRemaining: 68,
      sans: [clean, `*.${clean}`],
    },
    subdomains: [
      { subdomain: `www.${clean}`, ips: ['104.21.44.19'], cname: `${clean}.cdn.cloudflare.net`, takeoverRisk: 'Safe (Active IP Resolution)' },
      { subdomain: `docs.${clean}`, ips: ['104.21.44.19'], takeoverRisk: 'Safe (Active IP Resolution)' },
      { subdomain: `api.${clean}`, ips: ['172.67.188.91'], takeoverRisk: 'Safe (Active IP Resolution)' },
    ],
    httpHeaders: {
      Server: 'cloudflare',
      'Strict-Transport-Security': 'max-age=31536000; includeSubDomains; preload',
      'X-Content-Type-Options': 'nosniff',
      'Referrer-Policy': 'strict-origin-when-cross-origin',
    },
    hasRobotsTxt: true,
    hasSecurityTxt: true,
    hasSitemapXml: true,
    securityGrade: 'A',
    securityScore: 85,
    securityChecks: [
      { control: 'Strict-Transport-Security (HSTS)', passed: true, detail: 'Enforces encrypted HTTPS connections' },
      { control: 'Content-Security-Policy (CSP)', passed: true, detail: 'Active XSS & injection policy header' },
      { control: 'Clickjacking Defense (X-Frame / Ancestors)', passed: true, detail: 'Frame embedding restricted' },
      { control: 'DNS Sender Policy Framework (SPF)', passed: true, detail: 'Authorized outbound mail servers declared' },
      { control: 'DNS DMARC Anti-Spoofing Policy', passed: false, detail: 'No _dmarc record published' },
    ],
    trackers: {
      privacyGrade: 'A',
      verdict: 'Minimal First-Party Analytics Only',
      summary: 'Clean ad-free surface using 1 analytics tool (Cloudflare Web Analytics) with zero behavioral retargeting pixels.',
      adNetworksCount: 0,
      analyticsCount: 1,
      pixelsCount: 0,
      telemetryCount: 0,
      totalDetected: 1,
      detectedTrackers: [
        {
          name: 'Cloudflare Web Analytics',
          category: 'ANALYTICS',
          provider: 'Cloudflare',
          matchedRule: 'static.cloudflareinsights.com/beacon.min.js',
          riskLevel: 'LOW',
          description: 'Privacy-first cookieless edge performance & visitor beacon',
        },
      ],
    },
    durationMs: 310,
  };
}

function synthesizeCrawlReport(rawUrl: string): CrawlReport {
  const clean = rawUrl.replace(/^https?:\/\//, '');
  const parts = clean.split('/');
  const host = parts[0] || 'svelte.dev';
  const seedPath = parts.length > 1 && parts.slice(1).join('/') ? '/' + parts.slice(1).join('/') : '/';

  return {
    rootUrl: `https://${host}${seedPath}`,
    seedPath,
    host,
    pagesCrawled: 6,
    totalLinks: 64,
    externalCount: 11,
    techHeaders: ['Server: Cloudflare', 'Platform: Vercel Edge'],
    durationMs: 412,
    pages: [
      { url: `https://${host}${seedPath}`, path: seedPath, title: `${host}${seedPath} — Primary Route`, h1: `Section ${seedPath}`, statusCode: 200, contentType: 'text/html', latencyMs: 42, depth: 0, internalLinks: 18, externalLinks: 4 },
      { url: `https://${host}/docs`, path: '/docs', title: `Documentation — ${host}`, h1: 'Introduction & Quickstart', statusCode: 200, contentType: 'text/html', latencyMs: 55, depth: 1, internalLinks: 14, externalLinks: 2 },
      { url: `https://${host}/docs/cli`, path: '/docs/cli', title: `CLI Reference — ${host}`, h1: 'Command Line Interface', statusCode: 200, contentType: 'text/html', latencyMs: 49, depth: 2, internalLinks: 9, externalLinks: 1 },
      { url: `https://${host}/docs/api`, path: '/docs/api', title: `REST & RDAP API — ${host}`, h1: 'Endpoints & Schemas', statusCode: 200, contentType: 'text/html', latencyMs: 61, depth: 2, internalLinks: 8, externalLinks: 1 },
      { url: `https://${host}/pricing`, path: '/pricing', title: `Plans & Enterprise — ${host}`, h1: 'Simple, Predictable Pricing', statusCode: 200, contentType: 'text/html', latencyMs: 38, depth: 1, internalLinks: 6, externalLinks: 0 },
      { url: `https://${host}/blog`, path: '/blog', title: `Engineering Blog — ${host}`, h1: 'Latest Releases & Architecture', statusCode: 200, contentType: 'text/html', latencyMs: 51, depth: 1, internalLinks: 9, externalLinks: 3 },
    ],
    tree: {
      segment: seedPath === '/' ? host : `${host}${seedPath}`,
      fullPath: seedPath,
      title: `${host}${seedPath}`,
      statusCode: 200,
      latencyMs: 42,
      internalOut: 18,
      children: [
        {
          segment: 'docs',
          fullPath: '/docs',
          title: 'Documentation Hub',
          statusCode: 200,
          latencyMs: 55,
          children: [
            { segment: 'cli', fullPath: '/docs/cli', title: 'CLI Reference', statusCode: 200, latencyMs: 49 },
            { segment: 'api', fullPath: '/docs/api', title: 'REST & RDAP API', statusCode: 200, latencyMs: 61 },
          ],
        },
        { segment: 'pricing', fullPath: '/pricing', title: 'Plans & Enterprise', statusCode: 200, latencyMs: 38 },
        { segment: 'blog', fullPath: '/blog', title: 'Engineering Blog', statusCode: 200, latencyMs: 51 },
      ],
    },
    trackers: {
      privacyGrade: 'A',
      verdict: 'Minimal First-Party Analytics Only',
      summary: 'Clean ad-free surface using 1 analytics tool (Cloudflare Web Analytics).',
      adNetworksCount: 0,
      analyticsCount: 1,
      pixelsCount: 0,
      telemetryCount: 0,
      totalDetected: 1,
      detectedTrackers: [
        {
          name: 'Cloudflare Web Analytics',
          category: 'ANALYTICS',
          provider: 'Cloudflare',
          matchedRule: 'static.cloudflareinsights.com/beacon.min.js',
          riskLevel: 'LOW',
          description: 'Privacy-first cookieless edge performance & visitor beacon',
        },
      ],
    },
  };
}

function synthesizeRobotsSitemapReport(rawTarget: string): RobotsSitemapReport {
  const host = rawTarget.replace(/^https?:\/\//, '').split('/')[0] || 'svelte.dev';
  return {
    targetUrl: `https://${host}`,
    host,
    checkedAt: new Date().toISOString(),
    durationMs: 245,
    robotsFound: true,
    robotsUrl: `https://${host}/robots.txt`,
    robotsStatus: 200,
    robotsSizeBytes: 512,
    totalDisallowCount: 4,
    totalAllowCount: 2,
    declaredSitemaps: [`https://${host}/sitemap.xml`],
    agentGroups: [
      { userAgent: '*', disallow: ['/admin/', '/api/private/', '/checkout/'], allow: ['/'] },
      { userAgent: 'GPTBot', disallow: ['/'], allow: [] },
    ],
    botMatrix: [
      { botName: 'Googlebot', category: 'SEARCH_ENGINE', status: 'PARTIAL', matchedRule: 'User-agent: * → Disallow: /admin/ (+2 paths)' },
      { botName: 'Bingbot', category: 'SEARCH_ENGINE', status: 'PARTIAL', matchedRule: 'User-agent: * → Disallow: /admin/ (+2 paths)' },
      { botName: 'GPTBot (OpenAI)', category: 'AI_LLM', status: 'BLOCKED', matchedRule: 'User-agent: GPTBot → Disallow: /' },
      { botName: 'ClaudeBot (Anthropic)', category: 'AI_LLM', status: 'PARTIAL', matchedRule: 'User-agent: * → Disallow: /admin/' },
      { botName: 'Twitterbot (X Card)', category: 'SOCIAL_PREVIEW', status: 'ALLOWED', matchedRule: 'Allowed (Card previews permitted)' },
      { botName: 'facebookexternalhit', category: 'SOCIAL_PREVIEW', status: 'ALLOWED', matchedRule: 'Allowed (OpenGraph previews permitted)' },
    ],
    rawRobotsPreview: `User-agent: *\nAllow: /\nDisallow: /admin/\nDisallow: /api/private/\n\nUser-agent: GPTBot\nDisallow: /\n\nSitemap: https://${host}/sitemap.xml`,
    sitemapFound: true,
    sitemapUrl: `https://${host}/sitemap.xml`,
    sitemapStatus: 200,
    isSitemapIndex: false,
    totalUrlsCount: 4,
    newestLastMod: '2026-09-20',
    oldestLastMod: '2025-11-04',
    updatedLast7Days: 2,
    updatedLast30Days: 3,
    updatedLastYear: 4,
    entries: [
      { loc: `https://${host}/`, path: '/', lastMod: '2026-09-20', ageLabel: '5d ago', changeFreq: 'daily', priority: '1.0' },
      { loc: `https://${host}/docs`, path: '/docs', lastMod: '2026-09-18', ageLabel: '7d ago', changeFreq: 'weekly', priority: '0.9' },
      { loc: `https://${host}/pricing`, path: '/pricing', lastMod: '2026-09-01', ageLabel: '24d ago', changeFreq: 'monthly', priority: '0.8' },
      { loc: `https://${host}/blog`, path: '/blog', lastMod: '2025-11-04', ageLabel: '10mo ago', changeFreq: 'weekly', priority: '0.7' },
    ],
  };
}

function synthesizeMetaSocialReport(rawTarget: string): MetaSocialReport {
  const clean = rawTarget.replace(/^https?:\/\//, '');
  const host = clean.split('/')[0] || 'svelte.dev';
  return {
    targetUrl: `https://${clean}`,
    finalUrl: `https://${clean}`,
    host,
    statusCode: 200,
    durationMs: 280,
    checkedAt: new Date().toISOString(),
    title: `${host} — Interactive Platform & Digital Studio`,
    description: `Explore ${clean} — curated tools, documentation, and interactive digital experiences.`,
    canonicalUrl: `https://${clean}`,
    faviconUrl: `https://${host}/favicon.ico`,
    themeColor: '#5366e8',
    ogTitle: `${host} — Interactive Platform & Digital Studio`,
    ogDescription: `Explore ${clean} — curated tools, documentation, and interactive digital experiences.`,
    ogSiteName: host,
    ogType: 'website',
    twitterCard: 'summary_large_image',
    resolvedTitle: `${host} — Interactive Platform & Digital Studio`,
    resolvedDescription: `Explore ${clean} — curated tools, documentation, and interactive digital experiences.`,
    resolvedSiteName: host,
    resolvedThemeColor: '#5366e8',
    socialScore: 80,
    socialGrade: 'A',
    auditChecks: [
      { id: 'title', label: 'HTML <title> Tag', status: 'PASS', details: '48 chars (Optimal 15–70 chars)' },
      { id: 'desc', label: 'Meta Description', status: 'PASS', details: '92 chars (Optimal 50–160 chars)' },
      { id: 'og_text', label: 'OpenGraph Title & Description', status: 'PASS', details: 'Configured for Discord/Telegram/WhatsApp' },
      { id: 'twitter_card', label: 'Twitter / Discord Card Mode', status: 'PASS', details: 'twitter:card = summary_large_image' },
    ],
    trackers: {
      privacyGrade: 'A',
      verdict: 'Minimal First-Party Analytics Only',
      summary: 'Clean ad-free surface using 1 analytics tool.',
      adNetworksCount: 0,
      analyticsCount: 1,
      pixelsCount: 0,
      telemetryCount: 0,
      totalDetected: 1,
      detectedTrackers: [],
    },
  };
}
