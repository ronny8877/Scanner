export interface Valuation {
  score: number;
  tier: 'Ultra Premium' | 'High Value' | 'Brandable' | 'Standard' | string;
  estimatedMinUsd: number;
  estimatedMaxUsd: number;
  estimatedDisplay: string;
  lengthScore: number;
  tldScore: number;
  phoneticScore: number;
  keywordScore: number;
  highlights: string[];
}

export interface ScanResultItem {
  domain: string;
  rootName: string;
  tld: string;
  available: boolean;
  status: 'Available' | 'Registered';
  registeredAt?: string;
  registrar?: string;
  nameservers?: string[];
  valuation: Valuation;
  latencyMs: number;
}

export interface ScanReport {
  seedKeywords: string[];
  totalChecked: number;
  availableCount: number;
  takenCount: number;
  highValueCount: number;
  durationMs: number;
  items: ScanResultItem[];
}

export interface DNSRecords {
  a?: string[];
  aaaa?: string[];
  mx?: string[];
  ns?: string[];
  txt?: string[];
  cname?: string;
}

export interface DomainInquiry {
  domain: string;
  available: boolean;
  statusSummary: string;
  registeredAt?: string;
  updatedAt?: string;
  expiresAt?: string;
  domainAge?: string;
  daysToExpiry?: number;
  registrar?: string;
  registryHandle?: string;
  statusFlags?: string[];
  nameservers?: string[];
  dns: DNSRecords;
  valuation: Valuation;
  checkedAt: string;
  checkLatencyMs: number;
}

export interface HistoryTimeline {
  date: string;
  source: string;
  event: string;
}

export interface HistoryReport {
  domain: string;
  previouslyRegistered: boolean;
  historyVerdict: string;
  summaryNote: string;
  firstSeenAt?: string;
  lastSeenAt?: string;
  firstSeenYear?: number;
  lastSeenYear?: number;
  totalSpanYears: number;
  waybackSnapshots: number;
  activeYears?: string[];
  certCount: number;
  pastIssuers?: string[];
  pastSubdomains?: string[];
  milestones?: HistoryTimeline[];
  checkLatencyMs: number;
}

export interface PortProbe {
  port: number;
  service: string;
  protocol: string;
  open: boolean;
  latencyMs: number;
  category: string;
  riskNote?: string;
}

export interface TLSInfo {
  supported: boolean;
  version?: string;
  cipherSuite?: string;
  issuer?: string;
  subject?: string;
  validFrom?: string;
  validUntil?: string;
  daysRemaining: number;
  sans?: string[];
}

export interface SubdomainHit {
  subdomain: string;
  ips: string[];
}

export interface SecurityCheck {
  control: string;
  passed: boolean;
  detail: string;
}

export interface ReconReport {
  domain: string;
  targetIp: string;
  openPortsCount: number;
  portsScanned: number;
  ports: PortProbe[];
  tls: TLSInfo;
  subdomains: SubdomainHit[];
  securityGrade: string;
  securityScore: number;
  securityChecks: SecurityCheck[];
  durationMs: number;
}

export interface PageInfo {
  url: string;
  path: string;
  title: string;
  description?: string;
  h1?: string;
  statusCode: number;
  contentType: string;
  latencyMs: number;
  depth: number;
  internalLinks: number;
  externalLinks: number;
  childrenPaths?: string[];
}

export interface SiteNode {
  segment: string;
  fullPath: string;
  title?: string;
  statusCode?: number;
  latencyMs?: number;
  internalOut?: number;
  externalOut?: number;
  children?: SiteNode[];
}

export interface CrawlReport {
  rootUrl: string;
  host: string;
  pagesCrawled: number;
  totalLinks: number;
  externalCount: number;
  techHeaders?: string[];
  durationMs: number;
  pages: PageInfo[];
  tree: SiteNode;
}

export interface SavedDomain {
  domain: string;
  available: boolean;
  status: string;
  valuation: Valuation;
  notes?: string;
  tags?: string[];
  previouslyRegistered?: boolean;
  firstSeenYear?: number;
  savedAt: string;
  lastCheckedAt: string;
}

export interface Job {
  id: string;
  type: 'scan' | 'inspect' | 'history' | 'recon' | 'crawl' | 'parallel_suite' | 'watchlist_recheck' | string;
  title: string;
  target: string;
  status: 'QUEUED' | 'RUNNING' | 'COMPLETED' | 'FAILED';
  progress: number;
  phase: string;
  workers: number;
  resultSummary: string;
  createdAt: string;
  durationMs: number;
  result?: unknown;
}

export interface ParallelSuiteResult {
  domain: string;
  inquiry: DomainInquiry;
  history: HistoryReport;
  recon: ReconReport;
  crawl: CrawlReport;
}
