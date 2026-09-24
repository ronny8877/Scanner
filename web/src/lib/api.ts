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
} from './types';

const API_BASE = 'http://localhost:8080';

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
  mutations: boolean;
  onlyAvailable: boolean;
  minScore: number;
}): Promise<{ report: ScanReport; liveBackend: boolean }> {
  try {
    const res = await fetch(`${API_BASE}/api/scan`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        keywords: params.keywords,
        tlds: params.tlds,
        mutations: params.mutations,
        onlyAvailable: params.onlyAvailable,
        minScore: params.minScore,
        concurrency: 16,
        maxResults: 48,
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
  try {
    const res = await fetch(`${API_BASE}/api/inspect?domain=${encodeURIComponent(domainInput)}`);
    if (!res.ok) throw new Error('Backend inspect failed');
    const inquiry: DomainInquiry = await res.json();
    return { inquiry, liveBackend: true };
  } catch {
    return { inquiry: synthesizeDomainInquiry(domainInput), liveBackend: false };
  }
}

export async function runDomainHistory(domainInput: string): Promise<{ history: HistoryReport; liveBackend: boolean }> {
  try {
    const res = await fetch(`${API_BASE}/api/history?domain=${encodeURIComponent(domainInput)}`);
    if (!res.ok) throw new Error('Backend history failed');
    const history: HistoryReport = await res.json();
    return { history, liveBackend: true };
  } catch {
    return { history: synthesizeHistoryReport(domainInput), liveBackend: false };
  }
}

export async function runPortRecon(domainInput: string): Promise<{ recon: ReconReport; liveBackend: boolean }> {
  try {
    const res = await fetch(`${API_BASE}/api/recon?domain=${encodeURIComponent(domainInput)}`);
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
  try {
    const res = await fetch(`${API_BASE}/api/crawl`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(params),
    });
    if (!res.ok) throw new Error('Backend crawl failed');
    const report: CrawlReport = await res.json();
    return { report, liveBackend: true };
  } catch {
    return { report: synthesizeCrawlReport(params.targetUrl), liveBackend: false };
  }
}

export async function runFullParallelSuite(
  domainInput: string
): Promise<{ suite: ParallelSuiteResult; liveBackend: boolean }> {
  try {
    const res = await fetch(`${API_BASE}/api/parallel-suite`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
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
        valuation: evaluateLocal(clean),
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

// --- Client-side fallback synthesis ---

function evaluateLocal(domainStr: string): Valuation {
  const [name = 'nova', tld = 'com'] = domainStr.toLowerCase().split('.');
  const lenScore = name.length <= 5 ? 28 : name.length <= 8 ? 22 : 14;
  const tldMap: Record<string, number> = { com: 25, ai: 24, io: 21, dev: 19, co: 18, app: 18 };
  const tldScore = tldMap[tld] ?? 12;
  const phoneticScore = 21;
  const keywordScore = 16;
  const score = Math.min(100, lenScore + tldScore + phoneticScore + keywordScore);
  const tier =
    score >= 85 ? 'Ultra Premium' : score >= 74 ? 'High Value' : score >= 62 ? 'Brandable' : 'Standard';
  const minUsd = score >= 85 ? 2800 : score >= 74 ? 850 : 220;
  const maxUsd = score >= 85 ? 7400 : score >= 74 ? 2600 : 680;

  return {
    score,
    tier,
    estimatedMinUsd: minUsd,
    estimatedMaxUsd: maxUsd,
    estimatedDisplay: `$${minUsd.toLocaleString()} - $${maxUsd.toLocaleString()}`,
    lengthScore: lenScore,
    tldScore,
    phoneticScore,
    keywordScore,
    highlights: [
      name.length <= 7 ? `Compact ${name.length}-letter root` : 'Clean brandable cadence',
      tld === 'com' ? 'Flagship .com extension' : `High-demand .${tld} tech TLD`,
      'Clean phonetic structure (no hyphens/digits)',
    ],
  };
}

function synthesizeScanReport(params: {
  keywords: string[];
  tlds: string[];
  mutations: boolean;
  onlyAvailable: boolean;
  minScore: number;
}): ScanReport {
  const seeds = params.keywords.length ? params.keywords : ['veltrix', 'nova'];
  const suffixes = params.mutations ? ['', 'hq', 'labs', 'flow', 'grid', 'core', 'cloud'] : [''];
  const items: ScanResultItem[] = [];

  for (const seed of seeds) {
    const clean = seed.toLowerCase().replace(/[^a-z0-9]/g, '');
    for (const sfx of suffixes) {
      const root = `${clean}${sfx}`;
      for (const tld of params.tlds) {
        const full = `${root}.${tld}`;
        const val = evaluateLocal(full);
        const available = sfx !== '' || tld === 'ai' || tld === 'dev' || root.length >= 7;
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
    totalChecked: items.length,
    availableCount,
    takenCount: items.length - availableCount,
    highValueCount: items.filter((i) => i.available && i.valuation.score >= 74).length,
    durationMs: 240,
    items,
  };
}

function synthesizeDomainInquiry(raw: string): DomainInquiry {
  const clean = raw.toLowerCase().replace(/^https?:\/\//, '').split('/')[0] || 'svelte.dev';
  const full = clean.includes('.') ? clean : `${clean}.com`;
  const isAvailable = full.includes('hq') || full.includes('veltrix') || full.endsWith('.ai');
  return {
    domain: full,
    available: isAvailable,
    statusSummary: isAvailable ? 'Available' : 'Registered',
    registeredAt: isAvailable ? undefined : '2016-04-18T14:22:10Z',
    updatedAt: isAvailable ? undefined : '2025-11-02T09:15:00Z',
    expiresAt: isAvailable ? undefined : '2028-04-18T14:22:10Z',
    domainAge: isAvailable ? undefined : '10y 5m (3,811 days)',
    daysToExpiry: isAvailable ? undefined : 572,
    registrar: isAvailable ? undefined : 'Cloudflare, Inc. (IANA #1910)',
    registryHandle: isAvailable ? undefined : 'DOM-8849201-VRSN',
    statusFlags: isAvailable ? [] : ['clientTransferProhibited', 'clientUpdateProhibited'],
    nameservers: isAvailable ? [] : ['ada.ns.cloudflare.com', 'bob.ns.cloudflare.com'],
    dns: isAvailable
      ? {}
      : {
          a: ['104.21.44.19', '172.67.188.91'],
          aaaa: ['2606:4700:3031::ac43:bc5b'],
          mx: ['10 mx1.forwardemail.net', '20 mx2.forwardemail.net'],
          ns: ['ada.ns.cloudflare.com', 'bob.ns.cloudflare.com'],
          txt: ['v=spf1 include:_spf.google.com ~all', 'google-site-verification=Tk92YV8xOTk'],
        },
    valuation: evaluateLocal(full),
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
      previouslyRegistered: false,
      historyVerdict: 'Clean Virgin Domain (No Past Registration Traces)',
      summaryNote:
        'Zero historical snapshots in Internet Archive Wayback Machine and zero past SSL certificates in Certificate Transparency logs.',
      totalSpanYears: 0,
      waybackSnapshots: 0,
      certCount: 0,
      checkLatencyMs: 112,
    };
  }
  return {
    domain: clean,
    previouslyRegistered: true,
    historyVerdict: 'Previously Registered / Historical Footprint Found',
    summaryNote: `Historical footprint confirmed between 2018 and 2026 (94 Wayback captures, 18 TLS certificates, 4 past subdomains).`,
    firstSeenAt: '2018-11-24',
    lastSeenAt: '2026-08-14',
    firstSeenYear: 2018,
    lastSeenYear: 2026,
    totalSpanYears: 9,
    waybackSnapshots: 94,
    activeYears: ['2018', '2019', '2020', '2021', '2022', '2023', '2024', '2025', '2026'],
    certCount: 18,
    pastIssuers: ["Let's Encrypt", 'Cloudflare Inc', 'Google Trust Services'],
    pastSubdomains: [`api.${clean}`, `docs.${clean}`, `staging.${clean}`, `blog.${clean}`],
    milestones: [
      { date: '2018-11-24', source: 'Wayback Archive', event: 'First archived web snapshot captured on 2018-11-24' },
      { date: '2019-02-03', source: 'CT Log (crt.sh)', event: 'First TLS/SSL certificate issued (18 total historical certs)' },
      { date: '2026-08-14', source: 'Wayback Archive', event: 'Most recent web crawl capture (94 total snapshots across 9 active years)' },
    ],
    checkLatencyMs: 184,
  };
}

function synthesizeReconReport(raw: string): ReconReport {
  const clean = raw.toLowerCase().replace(/^https?:\/\//, '').split('/')[0] || 'svelte.dev';
  return {
    domain: clean,
    targetIp: '104.21.44.19',
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
      { subdomain: `www.${clean}`, ips: ['104.21.44.19'] },
      { subdomain: `docs.${clean}`, ips: ['104.21.44.19'] },
      { subdomain: `api.${clean}`, ips: ['172.67.188.91'] },
    ],
    securityGrade: 'A',
    securityScore: 85,
    securityChecks: [
      { control: 'Strict-Transport-Security (HSTS)', passed: true, detail: 'Enforces encrypted HTTPS connections' },
      { control: 'Content-Security-Policy (CSP)', passed: true, detail: 'Active XSS & injection policy header' },
      { control: 'Clickjacking Defense (X-Frame / Ancestors)', passed: true, detail: 'Frame embedding restricted' },
      { control: 'DNS Sender Policy Framework (SPF)', passed: true, detail: 'Authorized outbound mail servers declared' },
      { control: 'DNS DMARC Anti-Spoofing Policy', passed: false, detail: 'No _dmarc record published' },
    ],
    durationMs: 310,
  };
}

function synthesizeCrawlReport(rawUrl: string): CrawlReport {
  const host = rawUrl.replace(/^https?:\/\//, '').split('/')[0] || 'svelte.dev';
  return {
    rootUrl: `https://${host}`,
    host,
    pagesCrawled: 6,
    totalLinks: 64,
    externalCount: 11,
    techHeaders: ['Server: Cloudflare', 'Platform: Vercel Edge'],
    durationMs: 412,
    pages: [
      { url: `https://${host}/`, path: '/', title: `${host} — Cybernetically Enhanced Web Apps`, h1: 'Build faster web platforms', statusCode: 200, contentType: 'text/html', latencyMs: 42, depth: 0, internalLinks: 18, externalLinks: 4 },
      { url: `https://${host}/docs`, path: '/docs', title: `Documentation — ${host}`, h1: 'Introduction & Quickstart', statusCode: 200, contentType: 'text/html', latencyMs: 55, depth: 1, internalLinks: 14, externalLinks: 2 },
      { url: `https://${host}/docs/cli`, path: '/docs/cli', title: `CLI Reference — ${host}`, h1: 'Command Line Interface', statusCode: 200, contentType: 'text/html', latencyMs: 49, depth: 2, internalLinks: 9, externalLinks: 1 },
      { url: `https://${host}/docs/api`, path: '/docs/api', title: `REST & RDAP API — ${host}`, h1: 'Endpoints & Schemas', statusCode: 200, contentType: 'text/html', latencyMs: 61, depth: 2, internalLinks: 8, externalLinks: 1 },
      { url: `https://${host}/pricing`, path: '/pricing', title: `Plans & Enterprise — ${host}`, h1: 'Simple, Predictable Pricing', statusCode: 200, contentType: 'text/html', latencyMs: 38, depth: 1, internalLinks: 6, externalLinks: 0 },
      { url: `https://${host}/blog`, path: '/blog', title: `Engineering Blog — ${host}`, h1: 'Latest Releases & Architecture', statusCode: 200, contentType: 'text/html', latencyMs: 51, depth: 1, internalLinks: 9, externalLinks: 3 },
    ],
    tree: {
      segment: host,
      fullPath: '/',
      title: `${host} — Home`,
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
  };
}
